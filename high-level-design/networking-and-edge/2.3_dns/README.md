# DNS — Domain Name System

DNS is the internet's distributed phone book: it translates human-readable hostnames (e.g., `api.example.com`) into IP addresses that routers understand. The translation is hierarchical and cached at every layer — from the OS resolver on your laptop to ISP-level recursive resolvers and root name servers — giving DNS both global scale and sub-millisecond lookup times for cached records. Because every client connection begins with a DNS lookup, DNS has evolved far beyond simple hostname resolution: it is now a core traffic-management layer supporting geographic routing, health-based failover, weighted canary releases, and a surprisingly effective form of global load balancing — all without touching a single application server.

---

## DNS Resolution Chain

```mermaid
sequenceDiagram
    autonumber
    participant App as Application
    participant Stub as Stub Resolver (OS)
    participant Rec as Recursive Resolver (ISP / 8.8.8.8)
    participant Root as Root Name Server
    participant TLD as TLD Name Server (.com)
    participant Auth as Authoritative NS (example.com)

    App->>Stub: resolve api.example.com
    Stub->>Rec: query (cache miss)

    Rec->>Root: who handles .com?
    Root-->>Rec: TLD NS for .com

    Rec->>TLD: who handles example.com?
    TLD-->>Rec: NS1.example.com, NS2.example.com

    Rec->>Auth: what is api.example.com?
    Auth-->>Rec: A record → 203.0.113.42, TTL=300

    Rec-->>Stub: 203.0.113.42 (cached for 300s)
    Stub-->>App: 203.0.113.42
```

**Key insight:** Only the first lookup pays the full chain cost. Subsequent lookups within the TTL window are served from resolver or OS cache in microseconds.

---

## DNS Record Types

```mermaid
flowchart TD
    Zone["DNS Zone\nexample.com"] --> A["A record\nHostname → IPv4\napi → 203.0.113.42"]
    Zone --> AAAA["AAAA record\nHostname → IPv6"]
    Zone --> CNAME["CNAME\nAlias → canonical name\nwww → example.com"]
    Zone --> MX["MX record\nMail exchanger"]
    Zone --> TXT["TXT record\nSPF, DKIM, verification"]
    Zone --> NS["NS record\nDelegates zone to name servers"]
    Zone --> SRV["SRV record\nService + port discovery\n_grpc._tcp → host:443"]

    classDef important fill:#2980B9,stroke:#1A5276,color:#fff
    classDef secondary fill:#7F8C8D,stroke:#616A6B,color:#fff

    class A,CNAME,NS important
    class AAAA,MX,TXT,SRV secondary
```

---

## TTL Trade-offs

```mermaid
flowchart LR
    subgraph Low["Low TTL (e.g. 30–60 s)"]
        direction TB
        LPro["+ Fast failover\n+ Easy deployments\n+ Quick A/B switches"]
        LCon["- Higher resolver load\n- More latency on first hit\n- Higher DNS bill"]
    end

    subgraph High["High TTL (e.g. 300–3600 s)"]
        direction TB
        HPro["+ Fewer queries\n+ Lower latency (more cache hits)\n+ Cheaper"]
        HCon["- Slow failover\n- Propagation lag during incidents\n- Stuck clients on IP change"]
    end

    classDef pro fill:#27AE60,stroke:#1E8449,color:#fff
    classDef con fill:#E74C3C,stroke:#C0392B,color:#fff

    class LPro,HPro pro
    class LCon,HCon con
```

| TTL Range | Use Case |
|---|---|
| 30–60 s | Active/active failover, canary deployments, on-call runbooks |
| 300 s (5 min) | Standard web services — good balance |
| 3600+ s (1 h+) | Static infrastructure, CDN origins, mail servers |
| 86400 s (1 day) | Root zone delegations, rarely-changed records |

---

## GeoDNS — Geographic Routing

```mermaid
flowchart TD
    User["User query\napi.example.com"]
    AuthNS["Authoritative NS\nwith GeoDNS"]

    User --> AuthNS

    AuthNS -->|"Resolver IP → US"| US["Return\n198.51.100.10\n(US region)"]
    AuthNS -->|"Resolver IP → EU"| EU["Return\n203.0.113.20\n(EU region)"]
    AuthNS -->|"Resolver IP → APAC"| APAC["Return\n192.0.2.30\n(APAC region)"]
    AuthNS -->|"Unknown / fallback"| FB["Return\n default/anycast IP"]

    US --> USRegion["US Load Balancer\n→ US servers"]
    EU --> EURegion["EU Load Balancer\n→ EU servers"]

    classDef ns fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef region fill:#2980B9,stroke:#1A5276,color:#fff

    class AuthNS ns
    class US,EU,APAC,FB region
```

**Caveat:** GeoDNS uses the resolver's IP (usually the ISP's), not the end user's. A user in Germany using Google's `8.8.8.8` resolver may be routed to the US PoP. EDNS Client Subnet (ECS) partially addresses this by forwarding a prefix of the client's IP to the authoritative server.

---

## Health-Based Routing (DNS Failover)

```mermaid
stateDiagram-v2
    [*] --> Primary_Active: initial state

    Primary_Active --> Failover_Check: health probe fails
    Failover_Check --> Secondary_Active: primary unhealthy\nDNS updated to secondary IP
    Secondary_Active --> Recovery_Check: primary probe passes
    Recovery_Check --> Primary_Active: failback (optional)\nor stay on secondary

    note right of Secondary_Active
        TTL must be low (30-60s)
        for fast propagation
    end note
```

| Provider | Health Check Frequency | Failover Time |
|---|---|---|
| AWS Route 53 | Every 10 or 30 s | ~60 s (TTL dependent) |
| Cloudflare DNS | Every 60 s | 2–3 min |
| Google Cloud DNS | External (Cloud Monitoring) | Minutes |
| Azure Traffic Manager | Every 30 s | ~3 min |

---

## DNS as a Load Balancer

DNS can distribute traffic without any hardware load balancer by returning multiple A records for the same hostname:

```
api.example.com.  60  IN  A  203.0.113.10
api.example.com.  60  IN  A  203.0.113.11
api.example.com.  60  IN  A  203.0.113.12
```

Clients (and resolvers) typically pick the first record, but many implementations rotate order on each response (round-robin DNS). Traffic is distributed but unevenly — stickier clients (browsers, OS caches) may always use the first address.

| Approach | Stickiness | Health Awareness | Use Case |
|---|---|---|---|
| Round-robin DNS | Low | No (TTL-based only) | Basic distribution, dev/staging |
| Weighted DNS (Route 53) | Configurable | Yes (with health checks) | Canary releases, blue/green |
| GeoDNS | By region | Yes | Multi-region active-active |
| Latency-based routing | By measured RTT | Yes | AWS Route 53 latency policy |
| Failover DNS | Active/passive | Yes | DR / standby environments |

---

## DNS Propagation Lag

DNS changes do not take effect instantly because cached records are served until their TTL expires:

```mermaid
flowchart TD
    Change["You update DNS record\nip: 10.0.0.1 → 10.0.0.2"] --> Published["Authoritative NS\nserves new record immediately"]
    Published --> RecResp["Recursive resolvers\nserve OLD record until their\ncache TTL expires"]
    RecResp --> Client["End clients\nsee new IP only after\nresolver cache expires"]

    Published -->|"TTL=300s"| Wait5["~5 min for most clients"]
    Published -->|"TTL=3600s"| Wait1h["Up to 1 hour"]
    Published -->|"TTL=86400s"| Wait24["Up to 24 hours"]

    classDef warning fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef ok fill:#27AE60,stroke:#1E8449,color:#fff

    class Wait1h,Wait24 warning
    class Wait5 ok
```

**Pre-migration checklist:** Lower TTL to 60 s at least 2x the current TTL before a planned migration. After migration, raise TTL back. This ensures a fast rollback window if something goes wrong.

---

## Trade-offs & When to Use

| Scenario | Recommendation |
|---|---|
| Global multi-region routing | GeoDNS + latency-based routing (Route 53 / Cloudflare) |
| Zero-downtime deployments | Low TTL (60 s) + weighted DNS for canary |
| Disaster recovery failover | DNS failover with health checks, passive secondary |
| Simple horizontal scaling | Round-robin DNS (no hardware LB needed) |
| CDN integration | CNAME to CDN hostname (CloudFront, Cloudflare) |
| Internal service discovery | Not DNS — use Consul, CoreDNS + Kubernetes, or Eureka |

---

## Real-World Examples

| Provider | Feature | Notes |
|---|---|---|
| AWS Route 53 | Latency, weighted, failover, geolocation policies | 100% SLA on authoritative NS |
| Cloudflare DNS | 1.1.1.1 resolver + anycast authoritative | Fastest public resolver (avg 11 ms) |
| Google Cloud DNS | Managed, 100% uptime SLA, low-latency anycast | Integrated with GCP load balancing |
| NS1 | Traffic steering with Filter Chain | Advanced: custom routing logic |
| Azure Traffic Manager | DNS-based LB with health probes | Works across Azure regions + on-prem |

---

## Interview Angles

**Q1: What happens when a DNS record's TTL expires mid-request?**

The TTL governs how long a cached record can be served without revalidation. The in-flight request at the moment of TTL expiry is not affected — it already has the IP and has opened a TCP connection. The next DNS lookup (next request, or next resolver cache expiry) will fetch a fresh record. There is no mid-connection renegotiation. This is why very short TTLs (sub-10 s) are rarely used in production — the performance cost of constant re-resolution outweighs the agility benefit.

**Q2: Why is GeoDNS not a perfect solution for geo-routing?**

GeoDNS resolves based on the **recursive resolver's IP**, not the end user's actual IP. Since many users share resolvers (ISP DNS, 8.8.8.8, 1.1.1.1), their traffic may be misrouted. Google's resolver in Mountain View would cause European users to receive US IPs. EDNS Client Subnet (ECS, RFC 7871) mitigates this by having the resolver include a masked prefix (e.g., /24) of the client's IP in the query, letting the authoritative server make a more accurate geo decision. Not all resolvers support ECS, and it has privacy implications.

**Q3: How would you design a zero-downtime DNS migration from one IP to another?**

1. Lower the record's TTL to 60 s at least `current_TTL * 2` time before the migration window (to drain resolver caches). 2. At migration time, update the A record to the new IP. 3. Monitor for exactly `TTL` seconds — old IP traffic will drop to near zero. 4. Decommission the old IP. 5. After confirming success, raise TTL back to 300 s or higher. If something goes wrong, revert immediately — with a 60 s TTL, clients see the rollback in about a minute. Never migrate with a multi-hour TTL; you lose the rollback window.

---

## Code

The Docker Compose setup in this module demonstrates round-robin DNS resolution using Docker network aliases — two services share the same hostname, and the Go client alternates between them:

```
high-level-design/networking-and-edge/2.3_dns/
├── web/
│   └── main.go          # Simple HTTP server reporting its container ID
├── client/
│   └── main.go          # Resolves hostname in a loop, shows which IP responds
├── docker-compose.yml   # web-1 and web-2 both registered as "web-service"
└── test.sh
```

Run it:

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```

**References**

- [RFC 1034 — Domain Names: Concepts and Facilities](https://datatracker.ietf.org/doc/html/rfc1034)
- [AWS Route 53 Routing Policies](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy.html)
- [Cloudflare — What is DNS?](https://www.cloudflare.com/learning/dns/what-is-dns/)
- [EDNS Client Subnet — RFC 7871](https://datatracker.ietf.org/doc/html/rfc7871)

---

## Architecture Diagrams

> Generated from D2 sources in [`docs/diagrams/src/2.3_dns/`](../../../docs/diagrams/src/2.3_dns/).
> Regenerate: `just diagrams-topic 2.3_dns`.

### Overview

The full resolution hierarchy: browser → recursive resolver → root → TLD → authoritative nameserver, with caches at every layer.

![Dns — overview](../../../docs/diagrams/svg/2.3_dns/overview/index.svg)

**`cache_hit` scenario:** the recursive resolver answers from cache and the root/TLD/authoritative chain never runs — the case for ~90% of queries.

![Dns — cache_hit scenario](../../../docs/diagrams/svg/2.3_dns/overview/cache_hit.svg)

**`ttl_expired` scenario:** the cached record ages out and the resolver re-walks the chain — TTL choice trades failover speed against query load.

![Dns — ttl_expired scenario](../../../docs/diagrams/svg/2.3_dns/overview/ttl_expired.svg)

### Request Flow

All 10 steps of an uncached lookup, each hop annotated with what it returns.

![Dns — request flow](../../../docs/diagrams/svg/2.3_dns/request-flow/index.svg)

**`dnssec` scenario:** each response carries signatures validated up the chain of trust, preventing spoofed answers (cache poisoning).

![Dns — dnssec scenario](../../../docs/diagrams/svg/2.3_dns/request-flow/dnssec.svg)
