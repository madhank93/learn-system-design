# 1.1 Monoliths vs Microservices

## What is a Monolith?

A **monolithic** application packages all business logic — UI, API, database access — into a single deployable unit. One codebase, one build, one process.

```
┌───────────────────────────────────────────────────┐
│                   Monolith                        │
│  ┌──────────┐  ┌──────────┐  ┌────────────────┐  │
│  │  User UI │  │ Order    │  │ Inventory      │  │
│  │          │  │ Service  │  │ Service        │  │
│  └──────────┘  └──────────┘  └────────────────┘  │
│               Single Process / Single DB           │
└───────────────────────────────────────────────────┘
```

## What are Microservices?

**Microservices** decompose the application into independently deployable services, each owning its data and communicating over the network (HTTP/gRPC/events).

```
┌───────────┐   HTTP    ┌───────────┐   Events  ┌───────────┐
│  User     │ ────────▶ │  Order    │ ─────────▶│ Inventory │
│  Service  │           │  Service  │           │  Service  │
│  (DB: pg) │           │  (DB: pg) │           │ (DB: redis│
└───────────┘           └───────────┘           └───────────┘
```

## Trade-offs

| Concern | Monolith | Microservices |
|---------|----------|---------------|
| **Simplicity** | Simple to develop and deploy | Complex: service mesh, discovery, distributed tracing |
| **Deployment** | Deploy everything on every change | Deploy only changed services |
| **Scaling** | Scale entire app (wasteful) | Scale hot services independently |
| **Data consistency** | Single ACID transaction | Distributed — need sagas/events |
| **Latency** | In-process function call (nanoseconds) | Network call (milliseconds) |
| **Failure isolation** | One bug crashes everything | Failures are isolated (circuit breaker needed) |
| **Team ownership** | One team owns everything (coordination hell at scale) | One team per service (Conway's Law) |
| **Testing** | Easy — run one process | Hard — need contract tests, mocks, integration suites |

## When to Choose What

**Start with a monolith when:**
- Team < 10 engineers
- Domain is not yet well understood
- MVP / early product
- Low operational complexity is valuable

**Move to microservices when:**
- Teams are blocked by each other's deployments
- A specific component needs drastically different scaling (e.g., 10x reads vs writes)
- Compliance requires data isolation between domains
- Different services need different tech stacks (rare)

## The Migration Path

Most successful microservice systems started as monoliths. The Strangler Fig pattern:

1. Identify bounded contexts (DDD)
2. Extract one service at a time, routing traffic via a proxy
3. Keep the monolith until extraction is complete
4. Never do a "big bang" rewrite

## Common Pitfall: Distributed Monolith

The worst outcome: split into many services that **must be deployed together** because they share a database or have tight synchronous coupling. You get the complexity of microservices with none of the benefits.

**Signs you have a distributed monolith:**
- Services share a database schema
- A change in Service A requires changes in Services B and C
- You can't deploy services independently

## References

1. [Microservices vs Monolith — Atlassian](https://www.atlassian.com/microservices/microservices-architecture/microservices-vs-monolith)
2. [Martin Fowler: Microservices](https://martinfowler.com/articles/microservices.html)
3. [Sam Newman: Building Microservices](https://www.oreilly.com/library/view/building-microservices-2nd/9781492034018/)
