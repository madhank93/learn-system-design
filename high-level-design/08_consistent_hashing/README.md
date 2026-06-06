## System Design HLD: Consistent Hashing

### Go ring router with 150 virtual nodes per physical node

Consistent hashing places nodes and keys on a circular hash ring. Adding or removing a node migrates only ~1/N keys (vs ~(N-1)/N for naive modulo hashing).

## Run

```sh
docker compose up --build -d
sleep 2
bash test.sh
```

## References
- Donne Martin, System Design Primer — consistent hashing section
- Alex Xu, "System Design Interview" — Chapter 5
