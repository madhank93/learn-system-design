## System Design HLD: Bloom Filter

### Probabilistic set membership — O(k) insert and lookup, zero false negatives

A bloom filter answers "is X in the set?" with:
- `definitely_no` — guaranteed correct
- `probably_yes` — may be a false positive

Space-efficient: 1000 items at 1% FP rate needs ~9.6 KB vs ~50 KB for a hash set.

## Run

```sh
docker compose up --build -d && sleep 1 && bash test.sh
```

## Test

```sh
go test -v ./server/
```

## References
- Burton H. Bloom, "Space/Time Trade-offs in Hash Coding" (1970)
- System Design Primer: Bloom filter use cases
