## System Design HLD: API Gateway

### nginx — path-based routing, API key auth, and versioning

| Port | Config | Feature |
|------|--------|---------|
| 8081 | routing | /api/users → user-service, /api/products → product-service |
| 8082 | auth | X-Api-Key validation (demo-key-123, demo-key-456) |
| 8083 | versioned | /v1/* deprecated, /v2/* current |

## Run

```sh
docker compose up --build -d
sleep 2
bash test.sh
```
