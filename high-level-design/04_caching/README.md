## System Design HLD: Caching

### nginx proxy_cache — no-cache baseline, TTL cache, and stale-while-revalidate

Caching stores upstream responses in nginx's on-disk cache. Subsequent requests for the same URI are served directly from cache without touching the backend.

## Configurations

| Port | Config | TTL | Stale |
|------|--------|-----|-------|
| 8081 | no cache | — | — |
| 8082 | proxy_cache | 30s | No |
| 8083 | stale | 10s | Yes |

## Run

```sh
docker compose up --build -d
bash test.sh
```

## References
- [nginx proxy_cache](https://nginx.org/en/docs/http/ngx_http_proxy_module.html)
- [nginx proxy_cache_use_stale](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_use_stale)
