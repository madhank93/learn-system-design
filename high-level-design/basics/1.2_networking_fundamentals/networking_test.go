package main

import (
	"testing"
	"time"
)

func TestLatencyTableOrdering(t *testing.T) {
	// Each tier should be slower than the previous
	for i := 1; i < len(LatencyTable); i++ {
		if LatencyTable[i].Latency < LatencyTable[i-1].Latency {
			t.Errorf("tier %d (%s=%v) faster than tier %d (%s=%v)",
				i, LatencyTable[i].Operation, LatencyTable[i].Latency,
				i-1, LatencyTable[i-1].Operation, LatencyTable[i-1].Latency)
		}
	}
}

func TestL1FasterThanRAM(t *testing.T) {
	var l1, ram time.Duration
	for _, t := range LatencyTable {
		switch t.Operation {
		case "L1 cache hit":
			l1 = t.Latency
		case "RAM read":
			ram = t.Latency
		}
	}
	if l1 >= ram {
		t.Fatalf("L1 (%v) should be faster than RAM (%v)", l1, ram)
	}
}

func TestRAMFasterThanSSD(t *testing.T) {
	var ram, ssd time.Duration
	for _, tier := range LatencyTable {
		switch tier.Operation {
		case "RAM read":
			ram = tier.Latency
		case "SSD random read":
			ssd = tier.Latency
		}
	}
	if ram >= ssd {
		t.Fatalf("RAM (%v) should be faster than SSD (%v)", ram, ssd)
	}
	// RAM should be at least 100x faster than SSD
	if ssd/ram < 100 {
		t.Errorf("expected SSD/RAM ratio >= 100x, got %dx", ssd/ram)
	}
}

func TestSameDCFasterThanCrossRegion(t *testing.T) {
	var samedc, crossregion time.Duration
	for _, tier := range LatencyTable {
		switch tier.Operation {
		case "Same datacenter round-trip":
			samedc = tier.Latency
		case "Cross-region round-trip":
			crossregion = tier.Latency
		}
	}
	if samedc >= crossregion {
		t.Fatalf("same-DC (%v) should be faster than cross-region (%v)", samedc, crossregion)
	}
}

func TestProtocolsHaveUseCases(t *testing.T) {
	for _, p := range Protocols {
		if p.Name == "" {
			t.Error("protocol has empty name")
		}
		if p.UseCase == "" {
			t.Errorf("protocol %q has no use case", p.Name)
		}
	}
}

func TestHTTP2IsMultiplexed(t *testing.T) {
	for _, p := range Protocols {
		if p.Name == "HTTP/2" && !p.Multiplexed {
			t.Error("HTTP/2 must be multiplexed")
		}
		if p.Name == "HTTP/1.1" && p.Multiplexed {
			t.Error("HTTP/1.1 must NOT be multiplexed")
		}
	}
}

func TestConnectionCostsExist(t *testing.T) {
	if len(ConnectionCosts) == 0 {
		t.Fatal("no connection cost entries")
	}
	for _, c := range ConnectionCosts {
		if c.Protocol == "" {
			t.Error("connection cost entry has empty protocol")
		}
		if c.RTTs < 0 {
			t.Errorf("negative RTTs for %s", c.Protocol)
		}
	}
}

func TestQUICLowerRTTThanTCPTLS12(t *testing.T) {
	var tcpTLS12RTTs, quicRTTs float64
	for _, c := range ConnectionCosts {
		switch c.Protocol {
		case "TCP + TLS 1.2":
			tcpTLS12RTTs = c.RTTs + c.TLSExtra
		case "QUIC (HTTP/3)":
			quicRTTs = c.RTTs + c.TLSExtra
		}
	}
	if quicRTTs >= tcpTLS12RTTs {
		t.Errorf("QUIC total RTTs (%.0f) should be less than TCP+TLS1.2 (%.0f)",
			quicRTTs, tcpTLS12RTTs)
	}
}
