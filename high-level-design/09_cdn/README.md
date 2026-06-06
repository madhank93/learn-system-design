## System Design HLD: CDN

### nginx edge cache — no-cache baseline, full edge cache, and selective caching

| Port | Config | Caches |
|------|--------|--------|
| 8081 | no cache | Nothing — every request hits origin |
| 8082 | edge_cache | All 200 responses, 5-min TTL |
| 8083 | selective | /static/* cached, /api/* bypassed |

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```

## References
- nginx proxy_cache documentation
- Donne Martin, System Design Primer — CDN section
