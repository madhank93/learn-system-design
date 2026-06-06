## System Design HLD: Retry + Exponential Backoff

### Go retry library with exponential backoff, max delay cap, and full jitter

Exponential backoff: delay doubles each attempt. Jitter: random value in [0, delay] prevents thundering herd when many clients retry simultaneously.

Default policy: 5 attempts, base=200ms, multiplier=2×, max=5s, jitter=on

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```

## Test

```sh
go test -v ./client/
```

## References
- AWS Architecture Blog: "Exponential Backoff and Jitter"
- Google SRE Book: "Handling Overload"
