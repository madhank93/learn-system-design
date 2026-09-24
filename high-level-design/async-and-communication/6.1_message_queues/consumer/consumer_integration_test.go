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
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("rabbitmq container: %v", err)
	}
	t.Cleanup(func() { c.Terminate(ctx) })
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5672")
	return fmt.Sprintf("amqp://guest:guest@%s:%s/", host, port.Port())
}

func TestIntegrationConsumeFromWorkQueue(t *testing.T) {
	url := startRabbitMQ(t)

	conn := dial(url)
	defer conn.Close()
	testCh, _ := conn.Channel()
	defer testCh.Close()

	q, _ := testCh.QueueDeclare("work-consume-test", false, true, false, false, nil)

	// Publish a message
	testCh.Publish("", q.Name, false, false, amqp.Publishing{
		Body: []byte("test-message"),
	})

	// Consume it
	testCh.Qos(1, 0, false)
	msgs, _ := testCh.Consume(q.Name, "", false, false, false, false, nil)

	select {
	case msg := <-msgs:
		if string(msg.Body) != "test-message" {
			t.Errorf("expected 'test-message', got %q", msg.Body)
		}
		msg.Ack(false)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestIntegrationFanoutDelivery(t *testing.T) {
	url := startRabbitMQ(t)

	conn := dial(url)
	defer conn.Close()
	testCh, _ := conn.Channel()
	defer testCh.Close()

	testCh.ExchangeDeclare("test-fanout", "fanout", false, true, false, false, nil)

	q1, _ := testCh.QueueDeclare("", false, true, true, false, nil)
	q2, _ := testCh.QueueDeclare("", false, true, true, false, nil)
	testCh.QueueBind(q1.Name, "", "test-fanout", false, nil)
	testCh.QueueBind(q2.Name, "", "test-fanout", false, nil)

	testCh.Publish("test-fanout", "", false, false, amqp.Publishing{
		Body: []byte("fanout-msg"),
	})

	msgs1, _ := testCh.Consume(q1.Name, "", true, false, false, false, nil)
	msgs2, _ := testCh.Consume(q2.Name, "", true, false, false, false, nil)

	received := 0
	for i := 0; i < 2; i++ {
		select {
		case <-msgs1:
			received++
		case <-msgs2:
			received++
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for fanout message")
		}
	}
	if received != 2 {
		t.Fatalf("expected 2 deliveries (one per queue), got %d", received)
	}
}
