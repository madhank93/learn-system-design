## System Design HLD: Circuit Breaker

### Go state machine — CLOSED / OPEN / HALF_OPEN with three threshold configs

| Port | Config | Threshold | Timeout |
|------|--------|-----------|---------|
| 8081 | sensitive | 2 failures | 3s |
| 8082 | standard | 3 failures | 5s |
| 8083 | tolerant | 5 failures | 10s |

Admin endpoints on port 8099: `/admin/fail` and `/admin/recover`

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```
