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

// registerService registers via Consul PUT API without a health check.
// Services without checks are immediately considered passing by Consul.
func registerService(t *testing.T, consulAddr, id, name, addr string, port int) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"ID":      id,
		"Name":    name,
		"Address": addr,
		"Port":    port,
	})
	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("http://%s/v1/agent/service/register", consulAddr),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("register failed: err=%v status=%d", err, resp.StatusCode)
	}
}

func TestIntegrationDiscoverFindsInstances(t *testing.T) {
	consulAddr := startConsul(t)

	// Register 2 instances without health checks — appear as passing immediately
	registerService(t, consulAddr, "svc-1", "test-svc", "10.0.0.1", 8081)
	registerService(t, consulAddr, "svc-2", "test-svc", "10.0.0.2", 8082)

	instances := discover(consulAddr, "test-svc")
	if len(instances) < 2 {
		t.Fatalf("expected at least 2 instances, got %d", len(instances))
	}
}

func TestIntegrationDiscoverEmpty(t *testing.T) {
	consulAddr := startConsul(t)
	instances := discover(consulAddr, "nonexistent-service")
	if len(instances) != 0 {
		t.Fatalf("expected 0 instances for unknown service, got %d", len(instances))
	}
}

func TestIntegrationDiscoverUnhealthyFiltered(t *testing.T) {
	consulAddr := startConsul(t)

	// Register with a health check pointing to an unreachable address inside
	// the Consul container's network — the check will fail, service won't pass.
	body, _ := json.Marshal(map[string]any{
		"ID":      "sick-svc",
		"Name":    "sick-service",
		"Address": "127.0.0.1",
		"Port":    19999,
		"Check": map[string]string{
			"HTTP":                           "http://127.0.0.1:19999/health",
			"Interval":                       "2s",
			"DeregisterCriticalServiceAfter": "30s",
		},
	})
	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("http://%s/v1/agent/service/register", consulAddr),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(req)

	// Wait for health check to run and fail
	time.Sleep(5 * time.Second)

	// discover uses ?passing=true — unhealthy service must not appear
	instances := discover(consulAddr, "sick-service")
	if len(instances) != 0 {
		t.Fatalf("unhealthy service should be filtered, got %d instances", len(instances))
	}
}
