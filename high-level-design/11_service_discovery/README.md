## System Design HLD: Service Discovery

### Consul self-registration with health checks and client-side round-robin

Services register themselves with Consul on startup. The client discovers healthy instances via Consul's health API and load-balances across them.

## Run

```sh
docker compose up --build -d
sleep 10
bash test.sh
```

## Consul UI

http://localhost:8500
