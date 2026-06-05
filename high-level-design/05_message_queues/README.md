+++
title = "Message Queues"
description = "Async decoupling, delivery guarantees, and pub/sub patterns — with Go and Docker examples."
weight = 4
sort_by = "weight"
taxonomies = { tags = ["Go", "Docker", "Distributed Systems"], categories = ["hld"] }

[extra]
toc_expand = true
+++

## What Is a Message Queue?

In a synchronous HTTP architecture, service A calls service B, waits for a response, then calls service C. Every hop adds latency and every downstream failure propagates upstream. If the order service is slow, the API gateway slows down. If the notification service crashes, the checkout fails. The services are tightly coupled in both time and failure domain.

A message queue breaks that coupling. The producer writes a message and moves on. The broker persists it. One or more consumers read it when they are ready. The producer does not know who processes the message, when, or how many times. This enables three things:

- **Async decoupling**: producer and consumer do not need to be alive at the same time.
- **Load leveling**: a traffic spike fills the queue instead of overwhelming downstream services.
- **Resilience**: if a consumer crashes, messages wait in the queue until it recovers.

{% mermaid() %}
graph LR
  subgraph Synchronous
    A1[Order Service] -->|HTTP| B1[Inventory Service]
    B1 -->|HTTP| C1[Notification Service]
  end

  subgraph Asynchronous
    A2[Order Service] -->|publish| Q[(Queue / Broker)]
    Q -->|consume| B2[Inventory Service]
    Q -->|consume| C2[Notification Service]
  end
{% end %}

In the async model, the order service publishes once. Inventory and notification services consume independently. A failure in notification does not affect inventory or the order service.

---

## Core Concepts

### Participants

**Producer** — the service that creates and publishes messages. It serializes the payload (JSON, Protobuf, Avro) and sends it to a named destination.

**Consumer** — the service that reads messages. It deserializes the payload, processes it, and acknowledges delivery.

**Broker** — the server in between. It receives messages from producers, stores them, and delivers them to consumers. Examples: RabbitMQ, Kafka, Redis Streams, AWS SQS.

### Delivery Patterns

**Queue (point-to-point)** — a single message is delivered to exactly one consumer. Multiple consumers compete for messages; each message is processed once. Used for task distribution: image resizing, email sending, invoice generation.

**Topic / Pub-Sub** — a single message is delivered to all subscribers. Each subscriber gets its own copy. Used for fan-out: audit logging, cache invalidation, analytics pipelines.

**Dead Letter Queue (DLQ)** — a secondary queue where messages are moved after exceeding a retry limit or failing validation. Prevents a bad message from blocking the main queue indefinitely. Consumed by an alerting or manual review process.

{% mermaid() %}
graph TD
  P[Producer] --> Q[(Queue)]
  Q -->|one consumer gets it| C1[Consumer A]
  Q -->|one consumer gets it| C2[Consumer B]
  Q -->|after N failures| DLQ[(Dead Letter Queue)]

  P2[Publisher] --> T[(Topic)]
  T -->|fan-out| S1[Subscriber 1]
  T -->|fan-out| S2[Subscriber 2]
  T -->|fan-out| S3[Subscriber 3]
{% end %}

---

## Delivery Guarantees

The contract between broker and consumer about what happens when things go wrong.

### At-Most-Once

The broker delivers the message once and forgets it. No retries. If the consumer crashes before processing, the message is lost. The lowest overhead — no acknowledgment round-trip, no state in the broker.

**Use when**: loss is acceptable — metrics aggregation, real-time dashboards, live game state. A missed data point does not matter.

### At-Least-Once

The broker retries delivery until it receives an acknowledgment. If the consumer processes the message but crashes before acking, the broker retries and the consumer processes it again. Duplicates are possible.

**Requirement**: consumers must be **idempotent** — processing the same message twice must produce the same result as processing it once.

**Use when**: correctness matters more than deduplication overhead — payment events, order processing, inventory updates (with idempotency keys).

### Exactly-Once

Each message is processed exactly once despite retries. This requires either distributed transactions (producer + broker + consumer in a single atomic unit) or idempotency keys checked by the consumer. The cost is high: 2-phase commit, transactional producers (Kafka), or deduplication state in the consumer.

**Use when**: duplicates have real business consequences and the overhead is justified — financial ledger entries, debit/credit operations.

{% mermaid() %}
sequenceDiagram
  participant P as Producer
  participant B as Broker
  participant C as Consumer

  P->>B: send(msg_id=42, payload)
  B->>C: deliver(msg_id=42)
  C->>C: process...
  Note over C: Consumer crashes before ack
  B-->>C: redeliver(msg_id=42) after timeout
  C->>C: process again (duplicate!)
  C->>B: ack(msg_id=42)
{% end %}

| Guarantee | Duplicates | Losses | Cost | Use Case |
|---|---|---|---|---|
| At-most-once | No | Possible | Low | Metrics, real-time events |
| At-least-once | Possible | No | Medium | Orders, payments (idempotent) |
| Exactly-once | No | No | High | Ledger entries, financial ops |

{% alert_warning() %}
Exactly-once is only as strong as the weakest link in your pipeline. A broker that supports exactly-once semantics cannot help you if your database write is not idempotent.
{% end %}

---

## Message Ordering

### Per-Partition Ordering

Kafka and similar log-based systems partition a topic into ordered, append-only logs. Messages within a single partition are guaranteed to be delivered in the order they were written. Messages across partitions are not ordered relative to each other.

Producers route messages to a partition by key (e.g., `user_id`, `order_id`). All events for a given key land in the same partition, ensuring ordered processing for that key.

### Global Ordering

Achievable with a single partition. The throughput ceiling is then one consumer. This is a bottleneck at scale — avoid it unless your use case genuinely requires total ordering across all keys.

{% mermaid() %}
graph LR
  P1[Producer A] -->|key=user_1| Part0[(Partition 0)]
  P1 -->|key=user_2| Part1[(Partition 1)]
  P2[Producer B] -->|key=user_3| Part2[(Partition 2)]

  Part0 --> CG1[Consumer Group\nInstance 1]
  Part1 --> CG2[Consumer Group\nInstance 2]
  Part2 --> CG3[Consumer Group\nInstance 3]
{% end %}

Each partition is assigned to exactly one consumer within a consumer group. Adding partitions and consumer instances scales throughput linearly, while preserving per-key ordering.

---

## Popular Systems Comparison

| System | Model | Persistence | Ordering | Throughput | Use Case |
|---|---|---|---|---|---|
| Redis Streams | Log / consumer groups | Optional (in-memory default) | Per-stream | Medium | Lightweight task queues, real-time pipelines |
| RabbitMQ | Traditional queue + topic | Durable (configurable) | Per-queue (FIFO) | Medium | Complex routing, AMQP, RPC-style patterns |
| Apache Kafka | Distributed log | Durable, replay-able | Per-partition | Very high | Event sourcing, audit logs, stream processing |
| AWS SQS | Managed queue | Durable | No ordering (FIFO queue option) | High | Managed workloads, serverless fan-out |
| AWS SNS | Managed pub-sub | No (push-only) | None | High | Fan-out to SQS, Lambda, HTTP endpoints |

**Key differentiator**: Kafka retains messages for a configurable period (days or weeks). Consumers can replay history. RabbitMQ and SQS delete messages after acknowledgment — there is no replay.

---

## Trade-offs and Failure Modes

| Concern | Issue | Mitigation |
|---|---|---|
| Consumer lag | Consumers process slower than producers publish | Add consumer instances, add partitions, apply backpressure to producers |
| Message loss | Broker crashes before flushing to disk | Durable queues, synchronous replication, `acks=all` in Kafka |
| Poison messages | Malformed message causes consumer to crash in a loop | DLQ with retry limit (e.g., 3–5 attempts), alert on DLQ depth |
| Duplicate processing | At-least-once + non-idempotent consumer | Idempotency keys checked in Redis or DB before processing |
| Out-of-order delivery | Network retries or partition rebalance reorder events | Sequence numbers per key, per-partition ordering |
| Queue depth explosion | Producers outpace consumers during a spike | Producer throttling, autoscaling consumer instances, circuit breakers |

- **Consumer lag** is the leading operational signal. Monitor it as a time value ("N seconds behind") not just a message count — a queue of 10,000 messages processed at 10k/s is fine; at 1/s it is a crisis.
- **Poison messages** are inevitable. Never retry infinitely. Always configure a DLQ and alert on it.
- **Duplicate processing** is the normal case with at-least-once. Design idempotent consumers from the start, not as a retrofit.

---

## Go Implementation: Simple Worker Queue

### In-Memory Queue with Graceful Shutdown

A buffered channel is Go's native bounded queue. It is sufficient for in-process job distribution, testing, and understanding queue semantics before introducing an external broker.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Job represents a unit of work.
type Job struct {
	ID      int
	Payload string
	Attempt int
}

const (
	maxRetries  = 3
	workerCount = 4
	queueSize   = 64
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	jobQueue := make(chan Job, queueSize)
	dlq := make(chan Job, queueSize) // Dead Letter Queue

	var wg sync.WaitGroup

	// Start workers.
	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker(ctx, workerID, jobQueue, dlq)
		}(i)
	}

	// Start DLQ drainer.
	wg.Add(1)
	go func() {
		defer wg.Done()
		drainDLQ(ctx, dlq)
	}()

	// Produce jobs.
	go func() {
		for i := 1; i <= 20; i++ {
			select {
			case <-ctx.Done():
				return
			case jobQueue <- Job{ID: i, Payload: fmt.Sprintf("task-%d", i)}:
				log.Printf("[producer] enqueued job %d", i)
			}
			time.Sleep(50 * time.Millisecond)
		}
		// Signal workers to stop after producing all jobs.
		// In production you would keep producing or use a separate shutdown signal.
		close(jobQueue)
	}()

	wg.Wait()
	log.Println("[main] shutdown complete")
}

// worker reads jobs from the queue, processes them, and sends failures to the DLQ.
func worker(ctx context.Context, id int, jobs <-chan Job, dlq chan<- Job) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[worker %d] context cancelled, stopping", id)
			return
		case job, ok := <-jobs:
			if !ok {
				log.Printf("[worker %d] queue closed, stopping", id)
				return
			}
			if err := process(job); err != nil {
				job.Attempt++
				if job.Attempt >= maxRetries {
					log.Printf("[worker %d] job %d failed after %d attempts → DLQ", id, job.ID, job.Attempt)
					dlq <- job
				} else {
					log.Printf("[worker %d] job %d failed (attempt %d), re-queuing", id, job.ID, job.Attempt)
					// Re-enqueue for retry (non-blocking to avoid deadlock).
					select {
					case jobs <- job:
					default:
						log.Printf("[worker %d] queue full, job %d → DLQ", id, job.ID)
						dlq <- job
					}
				}
			}
		}
	}
}

// process simulates work with a random failure rate.
func process(job Job) error {
	time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
	if rand.Float32() < 0.2 { // 20% failure rate
		return errors.New("transient error")
	}
	log.Printf("[process] job %d (%s) done", job.ID, job.Payload)
	return nil
}

// drainDLQ logs dead-lettered jobs; in production this would alert or persist them.
func drainDLQ(ctx context.Context, dlq <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-dlq:
			if !ok {
				return
			}
			log.Printf("[DLQ] job %d payload=%q attempts=%d — requires manual review",
				job.ID, job.Payload, job.Attempt)
		}
	}
}
```

Key properties of this implementation:

- `jobQueue` is a buffered channel — producers block only when the buffer is full, giving natural backpressure.
- Each worker selects on both the job channel and `ctx.Done()` so a SIGTERM drains gracefully.
- Failed jobs are re-enqueued up to `maxRetries`, then moved to `dlq`. This mimics broker-level DLQ semantics in a single process.
- The DLQ is a separate channel, making it easy to replace with an external queue (SQS, RabbitMQ) later.

### Docker Compose: RabbitMQ + Go Services

```yaml
services:
  rabbitmq:
    image: rabbitmq:3.13-management-alpine
    ports:
      - "5672:5672"    # AMQP
      - "15672:15672"  # Management UI
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - mqnet

  producer:
    build:
      context: ./producer
      dockerfile: Dockerfile
    environment:
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672/
      QUEUE_NAME: jobs
    depends_on:
      rabbitmq:
        condition: service_healthy
    networks:
      - mqnet

  consumer:
    build:
      context: ./consumer
      dockerfile: Dockerfile
    environment:
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672/
      QUEUE_NAME: jobs
    depends_on:
      rabbitmq:
        condition: service_healthy
    deploy:
      replicas: 2   # Scale consumers independently of producers
    networks:
      - mqnet

networks:
  mqnet:
    driver: bridge
```

The `condition: service_healthy` dependency prevents the Go services from starting before RabbitMQ is ready to accept connections — a common source of startup-race crashes.

---

## Idempotency Keys Pattern

At-least-once delivery is the practical default for most systems. That means your consumer **will** receive duplicates. The solution is not to hope for no duplicates — it is to make the consumer idempotent.

**Pattern**: before processing a message, check a fast store (Redis, Postgres) for the message ID. If it exists, skip. If not, process, then record it.

{% mermaid() %}
sequenceDiagram
  participant B as Broker
  participant C as Consumer
  participant R as Redis (idempotency store)
  participant DB as Database

  B->>C: deliver(msg_id=abc123)
  C->>R: EXISTS msg_id=abc123?
  R-->>C: NO
  C->>DB: write business logic
  C->>R: SET msg_id=abc123 EX 86400
  C->>B: ack(msg_id=abc123)

  Note over B,C: Broker retries (consumer crashed after write, before ack)

  B->>C: deliver(msg_id=abc123)
  C->>R: EXISTS msg_id=abc123?
  R-->>C: YES
  C->>B: ack(msg_id=abc123)
  Note over C: Duplicate skipped — DB write not repeated
{% end %}

Implementation notes:

- The idempotency key TTL should exceed your maximum message retention period. If messages can sit in the queue for 24 hours, keep idempotency keys for at least 48 hours.
- Use the broker-assigned message ID if it is stable across retries (RabbitMQ `delivery_tag` is not — use a custom header or UUID in the payload). Kafka offsets are stable.
- The check-then-set operation must be atomic. In Redis: `SET key 1 NX EX 86400`. In Postgres: `INSERT ... ON CONFLICT DO NOTHING`.

{% alert_info() %}
Design idempotency into your consumers from day one. Retrofitting it after a production duplicate-processing incident is painful and error-prone. Treat message delivery as "at least once" by default, even when your broker claims exactly-once, and you will never be surprised.
{% end %}

---

## How to Practice

- Run RabbitMQ locally with Docker (`docker run -p 5672:5672 -p 15672:15672 rabbitmq:management`) and write a producer/consumer pair in Go using the `amqp091-go` library. Observe the management UI as messages flow.
- Deliberately crash the consumer mid-processing and watch messages return to the queue. Then implement idempotency keys and verify duplicates are skipped.
- Set up a Kafka cluster with three brokers and two partitions using `docker-compose`. Produce 1000 messages keyed by a user ID and verify that all messages for the same user arrive in order within their partition.
- Implement a DLQ: configure a consumer that fails on messages containing the word "poison" and verify they appear in a separate DLQ after the retry limit.
- Benchmark throughput by measuring messages-per-second at different queue depths and consumer counts. Identify where the bottleneck is — producer, broker I/O, or consumer processing.
- Simulate consumer lag: produce messages faster than the consumer can process them. Observe queue depth grow. Then scale consumers and watch it drain.
- Read the Kafka consumer group rebalance protocol. Break a running consumer mid-partition and observe how the group coordinator reassigns partitions to surviving consumers.

---

## References

- "Designing Data-Intensive Applications" by Martin Kleppmann (Chapter 11: Stream Processing)
- Apache Kafka documentation — <https://kafka.apache.org/documentation/>
- RabbitMQ documentation — <https://www.rabbitmq.com/docs>
- System Design Primer by Donne Martin — <https://github.com/donnemartin/system-design-primer>
