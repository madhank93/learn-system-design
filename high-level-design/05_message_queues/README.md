## System Design HLD: Message Queues

### RabbitMQ direct queue (competing consumers) + fanout exchange (pub/sub)

A work queue distributes tasks across multiple competing consumers — each message is processed once. A fanout exchange broadcasts every message to all subscribers — each consumer gets its own copy.

## Run

```sh
docker compose up --build -d
sleep 5
bash test.sh
docker compose logs --tail=20 consumer-work-1 consumer-work-2
docker compose logs --tail=20 consumer-fanout-1 consumer-fanout-2
```

## Management UI

http://localhost:15672 — guest / guest

## References
- [RabbitMQ Tutorials](https://www.rabbitmq.com/tutorials)
