## System Design HLD: DNS

### Docker network aliases demonstrate round-robin DNS resolution

web-1 and web-2 both resolve to the Docker hostname `web-service`. The client alternates between them on each call.

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```
