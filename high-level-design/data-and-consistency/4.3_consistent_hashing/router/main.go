package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
)

const vnodes = 150

type ring struct {
	mu     sync.RWMutex
	hashes map[uint32]string
	sorted []uint32
}

func newRing() *ring { return &ring{hashes: make(map[uint32]string)} }

func (r *ring) hash(key string) uint32 {
	b := md5.Sum([]byte(key))
	return binary.BigEndian.Uint32(b[:4])
}

func (r *ring) add(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < vnodes; i++ {
		h := r.hash(fmt.Sprintf("%s#%d", node, i))
		r.hashes[h] = node
		r.sorted = append(r.sorted, h)
	}
	sort.Slice(r.sorted, func(i, j int) bool { return r.sorted[i] < r.sorted[j] })
	log.Printf("ring: added %s (%d vnodes)", node, vnodes)
}

func (r *ring) remove(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < vnodes; i++ {
		delete(r.hashes, r.hash(fmt.Sprintf("%s#%d", node, i)))
	}
	var filtered []uint32
	for _, h := range r.sorted {
		if _, ok := r.hashes[h]; ok {
			filtered = append(filtered, h)
		}
	}
	r.sorted = filtered
	log.Printf("ring: removed %s", node)
}

func (r *ring) get(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.sorted) == 0 {
		return ""
	}
	h := r.hash(key)
	idx := sort.Search(len(r.sorted), func(i int) bool { return r.sorted[i] >= h })
	if idx == len(r.sorted) {
		idx = 0
	}
	return r.hashes[r.sorted[idx]]
}

var gr *ring

func forward(w http.ResponseWriter, r *http.Request, node string) {
	url := fmt.Sprintf("http://%s%s", node, r.RequestURI)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("X-Node", node)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	gr = newRing()
	for _, n := range strings.Split(os.Getenv("NODES"), ",") {
		gr.add(strings.TrimSpace(n))
	}

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		node := gr.get(key)
		log.Printf("key:%s → %s", key, node)
		forward(w, r, node)
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		node := gr.get(key)
		log.Printf("key:%s → %s", key, node)
		forward(w, r, node)
	})

	http.HandleFunc("/ring/add", func(w http.ResponseWriter, r *http.Request) {
		node := r.URL.Query().Get("node")
		gr.add(node)
		fmt.Fprintf(w, "added %s\n", node)
	})

	http.HandleFunc("/ring/remove", func(w http.ResponseWriter, r *http.Request) {
		node := r.URL.Query().Get("node")
		gr.remove(node)
		fmt.Fprintf(w, "removed %s\n", node)
	})

	log.Println("Ring router on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] router: %v", err)
	}
}
