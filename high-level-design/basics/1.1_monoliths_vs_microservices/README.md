# Monoliths vs Microservices

**Understanding the architectural spectrum between monoliths and microservices is foundational to every system design conversation.** The choice is not merely technical — it shapes team structure, deployment velocity, operational complexity, and failure blast radius. Monoliths excel early: one codebase, one database, no network hops, simple transactions. Microservices unlock independent scaling, polyglot persistence, and organizational autonomy — but at the cost of distributed systems complexity. Getting this tradeoff wrong in either direction has real consequences: Netflix spent years carefully decomposing their monolith; Segment famously went the opposite direction and re-merged microservices back into a monolith when the overhead outweighed the benefits.

---

## Monolith Architecture

```mermaid
flowchart TD
    Client([Client / Browser]) --> LB["Load Balancer"]

    subgraph Monolith["Monolith — Single Deployable Unit"]
        direction TB
        UI["UI Layer\n(templates / views)"]
        BL["Business Logic\n(orders, users, payments)"]
        DAL["Data Access Layer\n(ORM / SQL)"]
        UI --> BL --> DAL
    end

    LB --> Monolith
    DAL --> DB[("Shared\nDatabase")]

    classDef client fill:#F39C12,stroke:#D68910,color:#fff
    classDef lb fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef mono fill:#27AE60,stroke:#1E8449,color:#fff
    classDef db fill:#8E44AD,stroke:#6C3483,color:#fff

    class Client client
    class LB lb
    class Monolith,UI,BL,DAL mono
    class DB db
```

In a monolith, all modules share an in-process call stack. A `PaymentService` calling an `InventoryService` is a function call — zero network latency, no serialization, and atomic transactions are trivial. The single deployment artifact means a one-step build and deploy.

---

## Microservices Architecture

```mermaid
flowchart TD
    Client([Client / Browser]) --> GW["API Gateway\n(auth, routing, rate-limit)"]

    GW --> US["User Service\n:8001"]
    GW --> OS["Order Service\n:8002"]
    GW --> PS["Payment Service\n:8003"]
    GW --> IS["Inventory Service\n:8004"]

    US --> UDB[("Users DB\nPostgres")]
    OS --> ODB[("Orders DB\nPostgres")]
    PS --> PDB[("Ledger DB\nPostgres")]
    IS --> IDB[("Inventory DB\nMongo")]

    OS -->|"event: order.created"| MQ["Message Queue\n(Kafka)"]
    MQ -->|"consume"| PS
    MQ -->|"consume"| IS

    classDef client fill:#F39C12,stroke:#D68910,color:#fff
    classDef gw fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef svc fill:#27AE60,stroke:#1E8449,color:#fff
    classDef db fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef mq fill:#E74C3C,stroke:#C0392B,color:#fff

    class Client client
    class GW gw
    class US,OS,PS,IS svc
    class UDB,ODB,PDB,IDB db
    class MQ mq
```

Each service owns its data store and exposes a versioned API contract. Services communicate over HTTP/gRPC or asynchronously via a message bus. A deploy of `Payment Service` has zero risk to `User Service`.

---

## Monolith vs Microservices — Comparison

| Dimension | Monolith | Microservices |
|---|---|---|
| **Deployment unit** | Single artifact (JAR / binary) | One artifact per service |
| **Transaction model** | Local ACID transactions trivial | Distributed transactions (Saga / 2PC) |
| **Network hops** | Zero (in-process calls) | One or more per cross-service call |
| **Scaling granularity** | Scale the whole app | Scale individual services independently |
| **Team structure** | Single team can own everything | Conway's Law: one team per service |
| **Technology stack** | Uniform (one language/framework) | Polyglot (each team picks best tool) |
| **Operational complexity** | Low (one service to monitor/debug) | High (distributed tracing, many dashboards) |
| **Blast radius of failure** | Large (one bug can take down app) | Contained (circuit breakers isolate failures) |
| **Startup cost** | Low — ship fast on day one | High — infra overhead before first feature |
| **Refactoring** | Easy — compiler enforces contracts | Hard — API versioning, backward compat |
| **Testing** | Integration tests are simpler | E2E tests across services are brittle |
| **Data consistency** | Enforced by DB constraints | Eventual consistency is the norm |

---

## The Distributed Monolith Anti-Pattern

The worst outcome: services that look like microservices but behave like a monolith. You pay all the distributed systems costs with none of the benefits.

```mermaid
flowchart TD
    subgraph DistMono["Distributed Monolith — Anti-Pattern"]
        direction LR
        A["Service A"] -->|"sync HTTP\n(blocks)"| B["Service B"]
        B -->|"sync HTTP\n(blocks)"| C["Service C"]
        C -->|"sync HTTP\n(blocks)"| D["Service D"]
        A -.->|"shared DB schema"| SharedDB[("Shared DB\n(all services)")]
        B -.->|"shared DB schema"| SharedDB
        C -.->|"shared DB schema"| SharedDB
        D -.->|"shared DB schema"| SharedDB
    end

    classDef svc fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef db fill:#7F8C8D,stroke:#566573,color:#fff
    classDef box fill:#FADBD8,stroke:#E74C3C,color:#333

    class A,B,C,D svc
    class SharedDB db
```

**Signs you have a distributed monolith:**
- Services share a database schema or call each other's internal tables
- Deploying Service A requires coordinated deploys of B, C, and D
- A failure in Service D cascades synchronously to A (no circuit breaker)
- You cannot run any single service in isolation for testing

---

## The Strangler Fig Migration Pattern

Safely migrating a monolith to microservices without a risky "big bang" rewrite:

```mermaid
flowchart TD
    subgraph Phase1["Phase 1 — Facade in Front"]
        C1([Client]) --> F1["Facade / Proxy"]
        F1 -->|"all traffic"| M1["Monolith"]
    end

    subgraph Phase2["Phase 2 — Peel Off First Service"]
        C2([Client]) --> F2["Facade / Proxy"]
        F2 -->|"/payments/*"| PS2["Payment\nMicroservice"]
        F2 -->|"everything else"| M2["Monolith\n(payments removed)"]
    end

    subgraph Phase3["Phase 3 — Incrementally Replace"]
        C3([Client]) --> F3["Facade / Proxy"]
        F3 -->|"/payments/*"| PS3["Payment\nMicroservice"]
        F3 -->|"/users/*"| US3["User\nMicroservice"]
        F3 -->|"/orders/*"| OS3["Order\nMicroservice"]
        F3 -->|"legacy routes"| M3["Monolith\n(shrinking)"]
    end

    subgraph Phase4["Phase 4 — Monolith Deleted"]
        C4([Client]) --> F4["API Gateway"]
        F4 --> PS4["Payment Svc"]
        F4 --> US4["User Svc"]
        F4 --> OS4["Order Svc"]
        F4 --> NS4["Notification Svc"]
    end

    classDef client fill:#F39C12,stroke:#D68910,color:#fff
    classDef facade fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef mono fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef svc fill:#27AE60,stroke:#1E8449,color:#fff

    class C1,C2,C3,C4 client
    class F1,F2,F3,F4 facade
    class M1,M2,M3 mono
    class PS2,PS3,US3,OS3,PS4,US4,OS4,NS4 svc
```

**Key principle:** The facade (a reverse proxy or API gateway) routes requests. You never route the same path to both the monolith and the microservice simultaneously. Extract one bounded context at a time. Data migration is the hardest part — use the dual-write pattern (write to both old and new data stores, then cut over reads).

---

## Decision Framework

| Signal | Choose Monolith | Choose Microservices |
|---|---|---|
| **Team size** | < 8 engineers | > 15 engineers, multiple teams |
| **Domain complexity** | Well-understood, stable domain | Complex domain with clear bounded contexts |
| **Deployment frequency** | Weekly or less | Multiple times per day per service |
| **Scaling needs** | Uniform load across features | Different features have wildly different load |
| **Org structure** | Single team | Multiple autonomous product teams |
| **Stage** | Startup / MVP / new product | Mature product with proven product-market fit |
| **Operational maturity** | No dedicated platform/infra team | Strong SRE / DevOps culture, Kubernetes |
| **Compliance/isolation** | Not required | PCI, HIPAA require strict data isolation |

**Martin Fowler's rule:** Don't start with microservices. Start with a modular monolith, draw clear module boundaries, then extract services when you have real reasons to do so.

---

## Handling Distributed Transactions

When an operation spans multiple services (e.g., place order → charge payment → decrement inventory), there is no single database to give you ACID. The two main patterns:

```mermaid
flowchart TD
    subgraph Choreo["Choreography Saga"]
        direction LR
        OS_C["Order Service\nCREATE order"] -->|"order.created event"| PS_C["Payment Service\nCHARGE card"]
        PS_C -->|"payment.success event"| IS_C["Inventory Service\nDECREMENT stock"]
        IS_C -->|"inventory.reserved event"| NS_C["Notification Service\nSEND confirmation"]
        PS_C -->|"payment.failed event"| OS_C2["Order Service\nCANCEL order (compensate)"]
    end

    subgraph Orch["Orchestration Saga"]
        direction LR
        SAG["Saga\nOrchestrator"] -->|"1. charge"| PS_O["Payment Service"]
        PS_O -->|"2. success"| SAG
        SAG -->|"3. decrement"| IS_O["Inventory Service"]
        IS_O -->|"4. success"| SAG
        SAG -->|"5. notify"| NS_O["Notification Service"]
        SAG -->|"on any failure:\ncompensate"| COMP["Compensating\nTransactions"]
    end

    classDef svc fill:#27AE60,stroke:#1E8449,color:#fff
    classDef orch fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef comp fill:#E74C3C,stroke:#C0392B,color:#fff

    class OS_C,PS_C,IS_C,NS_C,OS_C2 svc
    class PS_O,IS_O,NS_O svc
    class SAG orch
    class COMP comp
```

| Approach | Pros | Cons |
|---|---|---|
| **Choreography** | Decoupled, no single point of failure | Hard to track overall saga state, event storms |
| **Orchestration** | Explicit flow, easy to visualize and debug | Orchestrator is a choke point, coupling |

---

## Real-World Examples

| Company | Journey | Key Insight |
|---|---|---|
| **Netflix** | Monolith → microservices (2008–2012) | Needed to scale streaming independently of DVD business; ~700 services today |
| **Amazon** | Monolith → service-oriented (2001–2006) | Bezos' "API mandate" forced every team to expose data via APIs |
| **Shopify** | Modular monolith (Ruby on Rails) | Deliberately resisted microservices; uses "pods" and feature flags instead |
| **Segment** | Microservices → monolith (2020) | 130 microservices became too complex; merged back into one Go service (Centrifuge) |
| **Uber** | Monolith → microservices → domain-oriented | Microservices sprawl → DOMA architecture to regroup services into domains |
| **Stack Overflow** | Monolith (still!) | 11M+ pageviews/day on a well-tuned .NET monolith; simplicity is underrated |

---

## Interview Angles

**Q1: How would you design a migration from a monolith to microservices?**
Use the Strangler Fig pattern — place a facade in front, extract one bounded context at a time (start with the highest-value or most independently-scalable domain), migrate data with dual-write, never do a big-bang rewrite.

**Q2: How do you handle distributed transactions across microservices?**
Use the Saga pattern — either choreography (event-driven, each service reacts) or orchestration (central coordinator directs steps). Compensating transactions handle rollback. For critical paths, consider the Outbox pattern to ensure at-least-once delivery.

**Q3: What is a distributed monolith and why is it worse than either extreme?**
A distributed monolith has microservice deployment units but tight runtime coupling — shared databases, synchronous call chains with no circuit breakers, coordinated deployments. You pay distributed systems overhead (network, serialization, partial failures) without gaining independent deployability or scaling.

**Q4: When should you NOT break a monolith into microservices?**
When the team is small (< 8 engineers), the domain is not yet understood (premature decomposition creates wrong service boundaries), or when there is no platform team to own container orchestration, distributed tracing, and service mesh. Operational overhead of microservices is non-trivial.

**Q5: How does Conway's Law relate to microservice boundaries?**
Conway's Law states that system architecture mirrors the communication structure of the organization. Microservice boundaries should match team boundaries — each team owns one or a few services end-to-end. If your org chart does not match your service graph, you get constant cross-team coordination (the distributed monolith anti-pattern).

---

**References**

- [Martin Fowler: Microservices](https://martinfowler.com/articles/microservices.html)
- [Martin Fowler: Strangler Fig Application](https://martinfowler.com/bliki/StranglerFigApplication.html)
- [Sam Newman: Building Microservices (O'Reilly)](https://samnewman.io/books/building_microservices/)
- [Segment: Goodbye Microservices](https://segment.com/blog/goodbye-microservices/)
- [Shopify: Deconstructing the Monolith](https://shopify.engineering/deconstructing-monolith-designing-software-maximizes-developer-productivity)

---

## Architecture Diagrams

> Generated from D2 sources in [`docs/diagrams/src/1.1_monoliths_vs_microservices/`](../../../docs/diagrams/src/1.1_monoliths_vs_microservices/).
> Regenerate: `just diagrams-topic 1.1_monoliths_vs_microservices`.

### Overview

Side-by-side: the monolith routes everything through one process sharing one database, while microservices give each domain its own service and datastore behind an API gateway.

![Monoliths Vs Microservices — overview](../../../docs/diagrams/svg/1.1_monoliths_vs_microservices/overview/index.svg)

**`deploy_risk` scenario:** a bug in the Payment code forces a redeploy of the entire monolith (User and Order restart too), while the microservices side redeploys only the Payment service — User and Order keep serving traffic.

![Monoliths Vs Microservices — deploy_risk scenario](../../../docs/diagrams/svg/1.1_monoliths_vs_microservices/overview/deploy_risk.svg)

### Request Flow

The same user request traced through both architectures — one in-process call chain vs three network hops through the gateway.

![Monoliths Vs Microservices — request flow](../../../docs/diagrams/svg/1.1_monoliths_vs_microservices/request-flow/index.svg)

**`microservice_failure` scenario:** an OOM in payment logic kills the whole monolith process, taking every feature down; in the microservices column only the Payment service dies — browsing and ordering still work.

![Monoliths Vs Microservices — microservice_failure scenario](../../../docs/diagrams/svg/1.1_monoliths_vs_microservices/request-flow/microservice_failure.svg)
