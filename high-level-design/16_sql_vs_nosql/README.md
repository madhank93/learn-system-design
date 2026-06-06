## System Design HLD: SQL vs NoSQL

### SQLite normalized schema (JOIN) vs in-memory document store (denormalized)

| Aspect | SQL (port 8081) | NoSQL (port 8082) |
|--------|----------------|-------------------|
| Schema | Strict — users + orders tables | Flexible — any JSON fields |
| Relations | JOIN at query time | Embed at write time |
| Consistency | ACID transactions | Eventual (application-level) |
| Best for | Complex queries, strong consistency | High write throughput, flexible schema |

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```

## References
- Martin Kleppmann, "DDIA" — Chapter 2: Data Models and Query Languages
- AWS: When to use SQL vs NoSQL
