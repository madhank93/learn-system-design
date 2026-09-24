//go:build integration

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startRabbitMQ(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-alpine",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForListeningPort("5672/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("rabbitmq container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5672")
	return fmt.Sprintf("amqp://guest:guest@%s:%s/", host, port.Port())
}

func TestIntegrationPublishToWorkQueue(t *testing.T) {
	url := startRabbitMQ(t)

	conn := dial(url)
	defer conn.Close()

	testCh, err := conn.Channel()
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	defer testCh.Close()
	ch = testCh

	// Declare queue
	q, err := testCh.QueueDeclare("work", true, false, false, false, nil)
	if err != nil {
		t.Fatalf("queue declare: %v", err)
	}

	// Publish
	err = testCh.Publish("", "work", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         []byte("test-task"),
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Verify message in queue
	time.Sleep(100 * time.Millisecond)
	q2, _ := testCh.QueueInspect(q.Name)
	if q2.Messages != 1 {
		t.Fatalf("expected 1 message in queue, got %d", q2.Messages)
	}
}

func TestIntegrationPublishToFanout(t *testing.T) {
	url := startRabbitMQ(t)

	conn := dial(url)
	defer conn.Close()

	testCh, err := conn.Channel()
	if err != nil {
		t.Fatalf("channel: %v", err)
	}
	defer testCh.Close()
	ch = testCh

	// Declare fanout exchange
	testCh.ExchangeDeclare("events", "fanout", true, false, false, false, nil)

	// Bind two queues to the fanout
	q1, _ := testCh.QueueDeclare("fanout-test-1", false, true, false, false, nil)
	q2, _ := testCh.QueueDeclare("fanout-test-2", false, true, false, false, nil)
	testCh.QueueBind(q1.Name, "", "events", false, nil)
	testCh.QueueBind(q2.Name, "", "events", false, nil)

	// Publish to fanout
	testCh.Publish("events", "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("broadcast-event"),
	})

	time.Sleep(200 * time.Millisecond)

	// Both queues should have 1 message
	r1, _ := testCh.QueueInspect(q1.Name)
	r2, _ := testCh.QueueInspect(q2.Name)

	if r1.Messages != 1 {
		t.Errorf("fanout-test-1: expected 1 message, got %d", r1.Messages)
	}
	if r2.Messages != 1 {
		t.Errorf("fanout-test-2: expected 1 message, got %d", r2.Messages)
	}
}
