## System Design HLD: Distributed Locking

### Redis SET NX PX with compare-and-delete release (Lua script)

Only one worker can hold the lock at a time. Lock auto-expires (TTL) to prevent deadlocks if the holder crashes.

| Worker | Port | Behavior |
|--------|------|---------|
| worker-1 | 8081 | First acquires lock |
| worker-2 | 8082 | Gets 409 while lock held |
| worker-3 | 8083 | Gets 409 while lock held |

## Run

```sh
docker compose up --build -d && sleep 3 && bash test.sh
```

## References
- Redis documentation: SET NX PX
- Martin Kleppmann, "How to do distributed locking" (2016)
- Redis Redlock algorithm
