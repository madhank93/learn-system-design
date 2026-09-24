//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startConsul(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/consul:1.18",
		ExposedPorts: []string{"8500/tcp"},
		Cmd:          []string{"agent", "-dev", "-client=0.0.0.0"},
		WaitingFor:   wait.ForHTTP("/v1/status/leader").WithPort("8500/tcp").WithStartupTimeout(60 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("consul container: %v", err)
	}
	t.Cleanup(func() { c.Terminate(ctx) })
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "8500")
	return fmt.Sprintf("%s:%s", host, port.Port())
}

// putConsul registers a service directly via Consul API using PUT.
func putConsul(t *testing.T, consulAddr string, payload map[string]any) {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("http://%s/v1/agent/service/register", consulAddr),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("consul PUT register failed: err=%v status=%v", err, resp)
	}
}

func TestIntegrationServiceRegisters(t *testing.T) {
	consulAddr := startConsul(t)

	// Register without a health check — service appears immediately in catalog
	putConsul(t, consulAddr, map[string]any{
		"ID":      "test-svc-1",
		"Name":    "test-service",
		"Address": "127.0.0.1",
		"Port":    8080,
	})

	resp, err := http.Get(fmt.Sprintf("http://%s/v1/catalog/service/test-service", consulAddr))
	if err != nil {
		t.Fatalf("catalog query failed: %v", err)
	}
	defer resp.Body.Close()

	var services []map[string]any
	json.NewDecoder(resp.Body).Decode(&services)

	if len(services) == 0 {
		t.Fatal("expected service to be registered in Consul, got none")
	}
	if services[0]["ServiceName"] != "test-service" {
		t.Errorf("expected ServiceName=test-service, got %v", services[0]["ServiceName"])
	}
}

func TestIntegrationServiceDeregisters(t *testing.T) {
	consulAddr := startConsul(t)

	putConsul(t, consulAddr, map[string]any{
		"ID":      "deregister-svc",
		"Name":    "deregister-service",
		"Address": "127.0.0.1",
		"Port":    9090,
	})

	// Deregister
	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("http://%s/v1/agent/service/deregister/deregister-svc", consulAddr), nil)
	http.DefaultClient.Do(req)

	resp, _ := http.Get(fmt.Sprintf("http://%s/v1/catalog/service/deregister-service", consulAddr))
	defer resp.Body.Close()
	var services []map[string]any
	json.NewDecoder(resp.Body).Decode(&services)
	if len(services) != 0 {
		t.Fatalf("expected 0 services after deregister, got %d", len(services))
	}
}
