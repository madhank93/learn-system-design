## System Design HLD: WebSockets / SSE

### Go WebSocket echo server + SSE broadcast, proxied through nginx

| Protocol | Endpoint | Use case |
|----------|----------|---------|
| WebSocket | /ws | Bidirectional, full-duplex (chat, games) |
| SSE | /events | Server→client push (notifications, live feeds) |
| HTTP | /publish | Trigger broadcast from outside |

nginx config: WebSocket needs `Upgrade`/`Connection` header passthrough; SSE needs `proxy_buffering off`.

## Run

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```

## WebSocket test (interactive)
```sh
npx wscat -c ws://localhost:8080/ws
```

## References
- RFC 6455: WebSocket Protocol
- MDN: Server-Sent Events
- nginx WebSocket proxying docs
