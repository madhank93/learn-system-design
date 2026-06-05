+++
title = "API Gateway"
description = "Auth, routing, rate limiting, and aggregation — how API gateways work in microservices with Go and Docker."
weight = 5
sort_by = "weight"
taxonomies = { tags = ["Go", "Docker", "Distributed Systems", "Nginx"], categories = ["hld"] }

[extra]
toc_expand = true
+++

## What Is an API Gateway?

### The problem without a gateway

In a microservices system, each service owns a slice of the domain: users, orders, inventory, payments, notifications. Without a gateway, every client talks directly to every service it needs.

That creates an **N×M coupling problem**: N client types (web, mobile, third-party) × M services = a mesh of connections, each requiring the client to know hostnames, ports, and protocols of individual services. Cross-cutting concerns — authentication, rate limiting, logging — must be reimplemented in every service. A mobile client making a single screen load might need to fire five parallel HTTP calls and stitch the results together itself.

### How a gateway solves this

An API gateway is the single entry point for all client traffic. It terminates client connections, applies cross-cutting policies once, and forwards requests to downstream services. Services behind the gateway can evolve independently; clients see a stable, unified surface.

{% mermaid() %}
graph TD
    subgraph Without Gateway
        C1[Client] --> US1[Users Service :8001]
        C1 --> OS1[Orders Service :8002]
        C1 --> IS1[Inventory Service :8003]
        C1 --> PS1[Payments Service :8004]
        C1 --> NS1[Notifications :8005]
    end

    subgraph With Gateway
        C2[Client] --> GW[API Gateway :443]
        GW --> US2[Users Service]
        GW --> OS2[Orders Service]
        GW --> IS2[Inventory Service]
        GW --> PS2[Payments Service]
        GW --> NS2[Notifications]
    end
{% end %}

---

## Core Responsibilities

A request entering the gateway passes through a pipeline of concerns before reaching any upstream service.

{% mermaid() %}
graph LR
    REQ[Incoming Request] --> AUTH[Auth & AuthZ]
    AUTH --> RL[Rate Limiting]
    RL --> ROUTE[Routing]
    ROUTE --> TRANS[Request Transform]
    TRANS --> AGG[Aggregation / Fan-out]
    AGG --> LB[Load Balancing]
    LB --> CB[Circuit Breaker]
    CB --> US[Upstream Service]
    US --> OBS[Observability\nlog · trace · metric]
    OBS --> RESP[Response Transform]
    RESP --> CLIENT[Client]
{% end %}

### Routing

The gateway maps inbound requests to upstream services based on:

- **Path-based**: `/api/users/*` → users service, `/api/orders/*` → orders service
- **Host-based**: `api.acme.com` → external API, `internal.acme.com` → internal API
- **Header-based**: `X-API-Version: v2` routes to the v2 deployment; `X-Region: eu` routes to EU cluster

Path rewrites are common — the gateway may strip `/api/users` before forwarding so upstream services see clean paths.

### Authentication & Authorization

JWT validation, API key lookup, and OAuth token introspection happen once at the gateway. Downstream services receive a pre-validated identity as injected headers (`X-User-ID`, `X-User-Roles`). This keeps services lean: they trust the gateway and read headers rather than reimplementing crypto and token fetching.

### Rate Limiting

Per-client or per-endpoint quotas are enforced before requests hit upstream services. Common strategies:

- **Token bucket**: burst allowed up to a bucket size, refilled at a steady rate
- **Fixed window**: N requests per 60-second window per API key
- **Sliding window**: smoother than fixed window, avoids boundary spikes

Rate limit state is often stored in Redis for horizontal scaling of gateway instances.

### Request/Response Transformation

- **Header injection**: add `X-Request-ID`, `X-Forwarded-For`, strip sensitive client headers
- **Body transformation**: reshape JSON fields, rename keys for legacy upstream APIs
- **Protocol translation**: accept REST/JSON from clients, call gRPC upstreams — the gateway handles protobuf serialization

### Aggregation / BFF (Backend for Frontend)

A single client action often needs data from multiple services. Rather than requiring the client to fan out, the gateway (or a BFF layer on top of it) fans out internally, aggregates results, and returns a single response. This reduces client round-trips and keeps aggregation logic server-side.

### Load Balancing

The gateway distributes requests across multiple instances of each upstream service. Round-robin, least-connections, and consistent hashing are common. This overlaps with what a dedicated load balancer does, but the gateway has application-layer context (path, headers) to make smarter decisions.

### Circuit Breaking

When an upstream service is slow or returning errors, open the circuit: return a fast failure (503 or a cached fallback) instead of queuing requests that will also time out. This prevents one degraded service from consuming gateway threads and cascading failures across the system.

### Observability

Every request passes through the gateway, making it the natural place for:

- **Structured access logs**: method, path, upstream, latency, status code
- **Distributed trace propagation**: inject `traceparent` / `X-B3-TraceId` headers, export spans to Jaeger or Zipkin
- **Metrics**: request rate, error rate, p50/p95/p99 latency per route — expose as Prometheus metrics

---

## API Gateway vs Reverse Proxy vs Service Mesh

| Capability | Reverse Proxy (Nginx) | API Gateway | Service Mesh (Istio/Linkerd) |
|---|---|---|---|
| TLS termination | Yes | Yes | Yes (mTLS between services) |
| Path-based routing | Basic | Rich (regex, headers, methods) | Limited |
| Authentication | Manual config | First-class (JWT, OAuth, API keys) | Identity via mTLS, no app-layer auth |
| Rate limiting | Via module | Built-in | Not typical |
| Request transformation | Limited | Yes | No |
| Aggregation / BFF | No | Yes | No |
| Circuit breaking | No | Yes | Yes (via sidecar) |
| Traffic direction | North-south (ingress) | North-south (ingress) | East-west (service-to-service) |
| Observability | Access logs | Logs, traces, metrics | Deep telemetry via sidecar |
| Typical deployment | CDN edge, static sites | Public/partner API ingress | Internal service mesh |

**North-south traffic** is between external clients and the system boundary. An API gateway is the right tool here.

**East-west traffic** is between services inside the cluster. A service mesh (sidecar proxies like Envoy) handles mTLS, retries, and observability without touching application code.

You can use both: the gateway handles external ingress and app-layer policy; the service mesh handles internal service-to-service concerns. Avoid using the gateway for east-west traffic — it introduces unnecessary hops and a central bottleneck.

---

## BFF Pattern (Backend for Frontend)

### Why

A mobile app and a web dashboard need different response shapes. The mobile app wants a single compact payload with denormalized fields to minimize battery/bandwidth use. The web dashboard wants richer nested data for a complex UI. If both share one API shape, you either over-fetch (waste bandwidth) or force the client to make extra calls.

### How

Deploy a dedicated gateway per client type. Each BFF talks to the same downstream microservices but composes, filters, and reshapes responses for its specific client. BFF logic is thin — it aggregates and transforms, it does not own business logic.

{% mermaid() %}
graph TD
    MOB[Mobile Client] --> MBFF[Mobile BFF]
    WEB[Web Client] --> WBFF[Web BFF]

    MBFF --> US[Users Service]
    MBFF --> OS[Orders Service]

    WBFF --> US
    WBFF --> OS
    WBFF --> IS[Inventory Service]
    WBFF --> PS[Payments Service]
{% end %}

The Mobile BFF calls Users and Orders, strips unneeded fields, and returns a flat object. The Web BFF calls all four services, joins the data, and returns a rich nested object. Both BFFs are deployed independently and owned by their respective client teams.

---

## Trade-offs and Failure Modes

| Concern | Issue | Mitigation |
|---|---|---|
| Single point of failure | Gateway down = all services unreachable | HA deployment: multiple instances behind a Layer 4 LB; health checks with auto-restart |
| Latency addition | Every request gains one extra network hop | Keep gateway stateless and CPU-light; avoid synchronous DB calls in the hot path |
| Vendor lock-in | Managed gateways (AWS API Gateway, Kong Cloud) embed proprietary config syntax | Use open standards (OpenAPI for route definitions, OAuth 2.0 for auth); store config in code |
| Configuration drift | Routing rules in the gateway diverge from deployed services | GitOps: gateway config lives in the same repo as services; CI validates on every PR |
| Cascading failures | Slow upstream blocks gateway worker threads, starving other routes | Per-route timeouts, circuit breakers, bulkhead thread pools |
| Security boundary | Gateway receives external traffic; internal services must not be directly reachable | mTLS for gateway-to-service calls; internal services bound to loopback or private network only |

---

## Go Implementation: Minimal API Gateway

The gateway below routes `/api/users/` and `/api/orders/` to their respective upstream services, validates a Bearer token with HMAC-SHA256, and injects an `X-Request-ID` header on every proxied request.

```go
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
)

var hmacSecret []byte

func init() {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	hmacSecret = []byte(secret)
}

// validateToken performs a simple HMAC-SHA256 check.
// In production replace this with a full JWT library (e.g. golang-jwt/jwt).
func validateToken(token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	payload, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, hmacSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

func extractBearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	return strings.TrimPrefix(header, "Bearer "), true
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := extractBearerToken(r)
		if !ok || !validateToken(token) {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()
		r.Header.Set("X-Request-ID", reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r)
	})
}

func newReverseProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid upstream URL %q: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error for %s: %v", r.URL.Path, err)
		http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
	}
	return proxy
}

func main() {
	usersUpstream := os.Getenv("USERS_UPSTREAM")
	if usersUpstream == "" {
		usersUpstream = "http://users-service:8081"
	}
	ordersUpstream := os.Getenv("ORDERS_UPSTREAM")
	if ordersUpstream == "" {
		ordersUpstream = "http://orders-service:8082"
	}

	usersProxy := newReverseProxy(usersUpstream)
	ordersProxy := newReverseProxy(ordersUpstream)

	mux := http.NewServeMux()

	// Route /api/users/* → users service
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/users")
		usersProxy.ServeHTTP(w, r)
	})

	// Route /api/orders/* → orders service
	mux.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/orders")
		ordersProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Compose middleware: requestID → auth → routing
	handler := requestIDMiddleware(authMiddleware(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
```

{% alert_info() %}
The token format used here is `payload.HMAC(payload)`. For production, use a full JWT library such as `github.com/golang-jwt/jwt/v5` which handles RS256/ES256 signatures, expiry (`exp`), issuer (`iss`), and audience (`aud`) claims correctly.
{% end %}

### Docker Compose setup

The mock echo services return the request path and headers so you can verify routing and header injection without writing any service code.

```yaml
services:
  gateway:
    build: ./gateway
    ports:
      - "8080:8080"
    environment:
      - HMAC_SECRET=super-secret-key
      - USERS_UPSTREAM=http://users-service:8081
      - ORDERS_UPSTREAM=http://orders-service:8082
    depends_on:
      - users-service
      - orders-service
    networks:
      - gwnet

  users-service:
    image: ealen/echo-server:latest
    environment:
      - PORT=8081
    networks:
      - gwnet

  orders-service:
    image: ealen/echo-server:latest
    environment:
      - PORT=8082
    networks:
      - gwnet

networks:
  gwnet:
    driver: bridge
```

To generate a test token matching the HMAC scheme above:

```sh
payload="user123"
secret="super-secret-key"
sig=$(echo -n "$payload" | openssl dgst -sha256 -hmac "$secret" | awk '{print $2}')
token="$payload.$sig"

# Should route to users service and return 200
curl -H "Authorization: Bearer $token" http://localhost:8080/api/users/profile

# Missing token — should return 401
curl http://localhost:8080/api/orders/42
```

---

## JWT Validation at the Gateway

### Standard pattern

The gateway validates the JWT signature and expiry claim. On success it extracts claims (user ID, roles, tenant) and forwards them as request headers to the upstream service. Upstream services are inside the trust boundary — they read those headers without re-validating the token.

{% mermaid() %}
sequenceDiagram
    participant C as Client
    participant GW as API Gateway
    participant SVC as Upstream Service

    C->>GW: POST /api/orders/ Authorization: Bearer <jwt>
    GW->>GW: Verify JWT signature & exp
    alt Invalid or expired
        GW-->>C: 401 Unauthorized
    else Valid
        GW->>SVC: POST /orders/ X-User-ID: 42 X-User-Roles: admin
        SVC->>SVC: Read X-User-ID header (no re-validation)
        SVC-->>GW: 201 Created
        GW-->>C: 201 Created
    end
{% end %}

{% alert_warning() %}
Never trust `X-User-ID`, `X-User-Roles`, or any identity header that arrives from an **external client**. These headers must only be set by the gateway after successful JWT validation. If a client can bypass the gateway and reach a service directly, it can forge these headers and impersonate any user. Restrict upstream services to the gateway's network range and bind them to non-public interfaces.
{% end %}

### Implementation notes

- Use asymmetric keys (RS256 or ES256): the gateway holds the public key only; the identity provider signs with the private key. No shared secret to rotate across services.
- Cache the JWKS endpoint response with a TTL (e.g. 5 minutes) to avoid fetching public keys on every request.
- Reject tokens missing `exp`, `iss`, or `aud` even if the signature is valid — these claims prevent token reuse across environments.
- Pass the raw `sub` claim as `X-User-ID`; pass `roles` or `scope` as `X-User-Roles`. Keep the header set small — large header blocks increase HTTP/2 HPACK overhead.

---

## How to Practice

- **Build the gateway**: Extend the Go example above to add per-IP rate limiting using a token bucket backed by a Go `sync.Map`. Test it with `wrk` or `hey`.
- **Add JWT validation**: Replace the HMAC check with `golang-jwt/jwt`. Generate test tokens with `jwt-cli` or a small Go helper. Verify that expired tokens return 401.
- **Implement a BFF**: Write two thin HTTP handlers — one for a "mobile" client (returns minimal fields) and one for a "web" client (returns full nested data). Both handlers call the same downstream mock.
- **Inject distributed tracing**: Add `go.opentelemetry.io/otel` to the gateway. Propagate `traceparent` headers. Run Jaeger locally and verify the trace appears end-to-end.
- **Simulate circuit breaking**: Use `github.com/sony/gobreaker` to wrap the reverse proxy calls. Drop the upstream with `docker compose stop users-service` and confirm the gateway returns 503 immediately instead of timing out.
- **Benchmark latency overhead**: Measure p99 latency with the gateway vs. hitting the upstream directly using `hey -n 10000 -c 100`. The gateway should add less than 2 ms on localhost; investigate if it does not.
- **Read Kong or Envoy config**: Take a real Kong `kong.yaml` or an Envoy `bootstrap.yaml` from a public repo. Map each directive to the concepts above (routes, plugins, clusters). This demystifies production gateway configuration.

---

## References

1. Newman, Sam. *Building Microservices*, 2nd ed. O'Reilly Media, 2021.
2. Kong Inc. *Kong Gateway Documentation*. [https://docs.konghq.com/](https://docs.konghq.com/)
3. Amazon Web Services. *Amazon API Gateway Documentation*. [https://docs.aws.amazon.com/apigateway/](https://docs.aws.amazon.com/apigateway/)
4. Martin, Donne. *The System Design Primer*. GitHub. [https://github.com/donnemartin/system-design-primer](https://github.com/donnemartin/system-design-primer)
5. Netflix Tech Blog. "Open Sourcing Zuul 2." Netflix Engineering Blog. [https://netflixtechblog.com/open-sourcing-zuul-2-82ea476cb2b3](https://netflixtechblog.com/open-sourcing-zuul-2-82ea476cb2b3)
