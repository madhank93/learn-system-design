## System Design HLD: CAP Theorem

### Demonstrating CP vs AP behavior under network partition

CAP Theorem: a distributed system can guarantee at most 2 of Consistency, Availability, Partition Tolerance.

| Pair | Mode | Partition behavior |
|------|------|--------------------|
| cp-1 / cp-2 (8081/8082) | CP | Rejects writes when replica unreachable → consistent |
| ap-1 / ap-2 (8083/8084) | AP | Accepts writes even during partition → available but stale |

## Run

```sh
docker compose up --build -d && sleep 3 && bash test.sh
```

## References
- Martin Kleppmann, "DDIA" — Chapter 9: Consistency and Consensus
- Eric Brewer, "CAP Twelve Years Later" (IEEE Computer 2012)
