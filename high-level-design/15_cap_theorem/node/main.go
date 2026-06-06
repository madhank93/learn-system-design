package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type entry struct {
	Val     string `json:"val"`
	Version int64  `json:"version"`
}

var (
	mu          sync.RWMutex
	store       = map[string]entry{}
	version     int64
	partitioned int32 // 1 = partition active
	mode        string // "cp" or "ap"
	peerAddr    string
)

func replicate(key, val string, ver int64) error {
	if atomic.LoadInt32(&partitioned) == 1 {
		return fmt.Errorf("partition active")
	}
	body, _ := json.Marshal(map[string]any{"key": key, "val": val, "version": ver})
	resp, err := http.Post(
		fmt.Sprintf("http://%s/internal/replicate", peerAddr),
		"application/json", bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func main() {
	mode = os.Getenv("MODE")
	if mode == "" {
		mode = "ap"
	}
	peerAddr = os.Getenv("PEER_ADDR")
	nodeID := os.Getenv("NODE_ID")

	// Internal replication receiver
	http.HandleFunc("/internal/replicate", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Key     string `json:"key"`
			Val     string `json:"val"`
			Version int64  `json:"version"`
		}
		json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		existing := store[payload.Key]
		if payload.Version > existing.Version {
			store[payload.Key] = entry{Val: payload.Val, Version: payload.Version}
		}
		mu.Unlock()
		log.Printf("[%s] replicated key:%s v:%d", nodeID, payload.Key, payload.Version)
		w.WriteHeader(200)
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("val")
		ver := atomic.AddInt64(&version, 1)

		// CP: refuse write if replica unreachable
		if mode == "cp" {
			if err := replicate(key, val, ver); err != nil {
				atomic.AddInt64(&version, -1)
				http.Error(w, fmt.Sprintf("CP: write rejected — replica unreachable: %v", err), 503)
				return
			}
		}

		mu.Lock()
		store[key] = entry{Val: val, Version: ver}
		mu.Unlock()

		// AP: replicate best-effort (don't fail write on error)
		if mode == "ap" {
			go replicate(key, val, ver) //nolint
		}

		fmt.Fprintf(w, `{"node":%q,"key":%q,"val":%q,"version":%d,"mode":%q}`+"\n",
			nodeID, key, val, ver, mode)
		log.Printf("[%s/%s] set key:%s v:%d", nodeID, mode, key, ver)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		mu.RLock()
		e, ok := store[key]
		mu.RUnlock()
		if !ok {
			http.Error(w, "not found", 404)
			return
		}
		fmt.Fprintf(w, `{"node":%q,"key":%q,"val":%q,"version":%d}`+"\n",
			nodeID, key, e.Val, e.Version)
	})

	http.HandleFunc("/admin/partition", func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&partitioned) == 0 {
			atomic.StoreInt32(&partitioned, 1)
			fmt.Fprintln(w, "partition: ON  — nodes cannot reach each other")
			log.Printf("[%s] partition enabled", nodeID)
		} else {
			atomic.StoreInt32(&partitioned, 0)
			fmt.Fprintln(w, "partition: OFF — network restored")
			log.Printf("[%s] partition disabled", nodeID)
		}
	})

	http.HandleFunc("/admin/status", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		keys := len(store)
		mu.RUnlock()
		fmt.Fprintf(w, `{"node":%q,"mode":%q,"partitioned":%v,"keys":%d}`+"\n",
			nodeID, mode, atomic.LoadInt32(&partitioned) == 1, keys)
	})

	// Small delay on startup to let peers come up
	time.Sleep(500 * time.Millisecond)
	log.Printf("Node %s mode=%s peer=%s on :8080", nodeID, mode, peerAddr)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
