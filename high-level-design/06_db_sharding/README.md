## System Design HLD: Database Sharding

### Hash-based and range-based shard routers in Go

A shard router distributes keys across 3 in-memory shard services. Hash routing gives even distribution; range routing allows range-query efficiency at the cost of potential hot shards.

| Port | Strategy | Use case |
|------|----------|---------|
| 8081 | Hash (MD5 mod N) | Even write distribution |
| 8082 | Range (numeric bucket) | Range query efficiency |

## Run

```sh
docker compose up --build -d
sleep 2
bash test.sh
```

## References
- Alex Xu, "System Design Interview" — Chapter on key-value store
- Martin Kleppmann, "DDIA" — Chapter 6: Partitioning
