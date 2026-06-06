## System Design HLD: CQRS & Event Sourcing

### Redis Streams as event store + projector materializing read views

Commands append events to a Redis Stream (write side). A projector reads the stream and builds materialized views in Redis Hashes (read side).

| Service | Port | Role |
|---------|------|------|
| writer | 8081 | POST /command — appends events |
| reader | 8082 | GET /state/:id — serves materialized view |

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```
