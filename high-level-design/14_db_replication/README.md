## System Design HLD: Database Replication

### Go primary + 2 replicas + nginx read/write router

Writes go to primary. Replicas poll the primary's /log endpoint every 500ms. Nginx routes /write to primary and /read to replicas round-robin.

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```

## References
- Martin Kleppmann, "DDIA" — Chapter 5: Replication
