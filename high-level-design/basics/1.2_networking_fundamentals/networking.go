package main

import (
	"fmt"
	"net/http"
	"time"
)

// Protocol represents an application-layer protocol characteristic.
type Protocol struct {
	Name        string
	Multiplexed bool   // multiple requests over one connection
	Binary      bool   // binary framing vs text
	UseCase     string
}

var Protocols = []Protocol{
	{"HTTP/1.1", false, false, "Simple APIs, browsers (legacy)"},
	{"HTTP/2", true, true, "Modern APIs, gRPC, browser (default)"},
	{"HTTP/3 (QUIC)", true, true, "Mobile, lossy networks"},
	{"gRPC", true, true, "Internal microservice RPC"},
	{"WebSocket", true, false, "Real-time bidirectional (chat, games)"},
}

// LatencyTier models the latency of different storage/network operations.
// These are Jeff Dean's famous "Numbers Every Engineer Should Know" (approximate).
type LatencyTier struct {
	Operation string
	Latency   time.Duration
	Notes     string
}

var LatencyTable = []LatencyTier{
	{"L1 cache hit", 1 * time.Nanosecond, "CPU register-level speed"},
	{"L2 cache hit", 4 * time.Nanosecond, ""},
	{"L3 cache hit", 10 * time.Nanosecond, ""},
	{"RAM read", 100 * time.Nanosecond, "~100ns, 10x L3"},
	{"SSD random read", 100 * time.Microsecond, "~0.1ms, 1000x RAM"},
	{"Same datacenter round-trip", 500 * time.Microsecond, "~0.5ms LAN"},
	{"HDD seek + read", 10 * time.Millisecond, "~10ms (avoid in hot path)"},
	{"Cross-region round-trip", 150 * time.Millisecond, "US-East → US-West"},
	{"Cross-continent round-trip", 300 * time.Millisecond, "US → Europe"},
}

// HTTPLatencyProbe measures GET latency to a URL (for demonstration).
func HTTPLatencyProbe(url string) (time.Duration, int, error) {
	start := time.Now()
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	return time.Since(start), resp.StatusCode, nil
}

// PrintLatencyTable prints the latency reference table.
func PrintLatencyTable() {
	fmt.Printf("%-35s %-15s %s\n", "Operation", "Latency", "Notes")
	fmt.Println("──────────────────────────────────────────────────────────────────")
	for _, t := range LatencyTable {
		fmt.Printf("%-35s %-15s %s\n", t.Operation, t.Latency, t.Notes)
	}
}

// TCPHandshakeOverhead shows how many RTTs each protocol requires before data.
type ConnectionCost struct {
	Protocol  string
	RTTs      float64
	TLSExtra  float64
	Notes     string
}

var ConnectionCosts = []ConnectionCost{
	{"TCP (no TLS)", 1, 0, "SYN/SYN-ACK/ACK only"},
	{"TCP + TLS 1.2", 1, 2, "1 TCP + 2 TLS = 3 RTT total"},
	{"TCP + TLS 1.3", 1, 1, "1 TCP + 1 TLS = 2 RTT total"},
	{"QUIC (HTTP/3)", 1, 0, "Combines TCP+TLS: 1 RTT total (0-RTT resume)"},
}

func main() {
	fmt.Println("=== Latency Numbers Every Engineer Should Know ===")
	PrintLatencyTable()

	fmt.Println("\n=== Connection Setup Cost (RTTs before first byte) ===")
	fmt.Printf("%-20s %-8s %-10s %s\n", "Protocol", "TCP RTT", "TLS RTT", "Notes")
	fmt.Println("──────────────────────────────────────────────────────")
	for _, c := range ConnectionCosts {
		fmt.Printf("%-20s %-8.0f %-10.0f %s\n", c.Protocol, c.RTTs, c.TLSExtra, c.Notes)
	}

	fmt.Println("\n=== Protocol Comparison ===")
	fmt.Printf("%-20s %-12s %-8s %s\n", "Protocol", "Multiplexed", "Binary", "Best use")
	fmt.Println("────────────────────────────────────────────────────────────────")
	for _, p := range Protocols {
		fmt.Printf("%-20s %-12v %-8v %s\n", p.Name, p.Multiplexed, p.Binary, p.UseCase)
	}
}
