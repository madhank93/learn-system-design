## System Design HLD: Rate Limiting

### nginx limit_req — per-IP, per-API-key, and burst configurations

Rate limiting controls how many requests a client can make in a given time window. nginx's `limit_req` module implements a leaky-bucket algorithm: excess requests are either delayed, queued in a burst buffer, or rejected with 429.

## Configurations

| Port | Config | Rate | Burst |
|------|--------|------|-------|
| 8081 | per-IP | 5 req/s | 0 |
| 8082 | per-API-key | 10 req/s | 0 |
| 8083 | burst | 2 req/s | 10 |

## Run

```sh
docker compose up --build -d
bash test.sh
```

## References
- [nginx limit_req_module](https://nginx.org/en/docs/http/ngx_http_limit_req_module.html)
