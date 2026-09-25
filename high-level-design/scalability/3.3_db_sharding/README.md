# Database Sharding

**Sharding is horizontal partitioning of data across multiple independent database nodes** — each node owns a distinct subset of rows, called a shard. When a single database can no longer handle the write throughput, data volume, or query load of a system, sharding is the answer. A single PostgreSQL node can comfortably handle ~10,000 writes/second on commodity hardware; WhatsApp processes 100 billion messages per day — roughly 1.2 million/second. At that scale, you need hundreds of shards. Getting the shard key right is the most consequential architectural decision in a sharded system: a bad key creates hotspots that defeat the entire purpose; a good key distributes load so evenly that adding capacity is just spinning up another shard.

---

## Horizontal vs Vertical Partitioning

```mermaid
flowchart LR
    subgraph Vertical["Vertical Partitioning\n(split by column)"]
        direction TB
        VT["users table\n(id, name, email, bio,\nprofile_pic_url, settings_json)"]
        VT --> VT1["Hot table\n(id, name, email)\nfrequently read"]
        VT --> VT2["Cold table\n(id, bio, settings_json)\nrarely read"]
    end

    subgraph Horizontal["Horizontal Partitioning / Sharding\n(split by row)"]
        direction TB
        HT["users table\n(1 billion rows)"]
        HT --> HS1["Shard 1\nusers 1–250M"]
        HT --> HS2["Shard 2\nusers 250M–500M"]
        HT --> HS3["Shard 3\nusers 500M–750M"]
        HT --> HS4["Shard 4\nusers 750M–1B"]
    end

    classDef vert fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef horiz fill:#2980B9,stroke:#1A5276,color:#fff
    classDef source fill:#F39C12,stroke:#D68910,color:#fff

    class VT source
    class VT1,VT2 vert
    class HT source
    class HS1,HS2,HS3,HS4 horiz
```

| Dimension | Vertical Partitioning | Horizontal Sharding |
|---|---|---|
| Split by | Columns | Rows |
| Addresses | Wide tables, I/O for unused columns | Volume and write throughput limits |
| Complexity | Low | High |
| Cross-partition queries | Require JOINs | Require scatter-gather |
| Typical tool | DB schema design, column families | Vitess, Cassandra, MongoDB |

---

## Shard Key Selection: Most Critical Decision

The shard key determines which shard receives each write and read. A poor choice means one shard receives 80% of traffic while others idle.

```mermaid
flowchart TD
    Data["Incoming Write\nuser_id=42, tweet content"] --> Router{Shard Router}

    Router -->|"Good key: hash(user_id)"| Even["Even Distribution\n● Shard 1: ~33%\n● Shard 2: ~33%\n● Shard 3: ~33%"]

    Router -->|"Bad key: timestamp"| Hot["Hot Spot\n● Shard 3 (latest): 100% writes\n● Shard 1, 2: idle (old data)"]

    Router -->|"Bad key: auto-increment ID"| Hot2["Hot Spot\n● Always writing to highest shard\n● Sequential writes = one active shard"]

    classDef good fill:#27AE60,stroke:#1E8449,color:#fff
    classDef bad fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef router fill:#2980B9,stroke:#1A5276,color:#fff
    classDef data fill:#F39C12,stroke:#D68910,color:#fff

    class Even good
    class Hot,Hot2 bad
    class Router router
    class Data data
```

---

## Sharding Strategies

### Hash-Based Sharding

```mermaid
flowchart TD
    Key["user_id = 12345"] --> Hash["hash(12345) = 0xA3F2\n→ 0xA3F2 % 3 = 1"]

    Hash --> S1["Shard 0\n(user_ids hashing to 0)"]
    Hash --> S2["Shard 1 ← user 12345\n(user_ids hashing to 1)"]
    Hash --> S3["Shard 2\n(user_ids hashing to 2)"]

    subgraph Pros["Hash Sharding Characteristics"]
        direction LR
        P1["✓ Even distribution\nno hot spots"]
        P2["✗ Range queries impossible\n(user 1–1000 spans all shards)"]
        P3["✗ Resharding requires\nmigrating ~50% of data"]
    end

    classDef shard fill:#2980B9,stroke:#1A5276,color:#fff
    classDef active fill:#27AE60,stroke:#1E8449,color:#fff
    classDef pro fill:#1ABC9C,stroke:#148F77,color:#fff
    classDef con fill:#E74C3C,stroke:#C0392B,color:#fff

    class S1,S3 shard
    class S2 active
    class P1,P2,P3 pro
```

### Range-Based Sharding

```mermaid
flowchart TD
    Key2["user_id = 1,500,000"] --> Range{Range Router}

    Range -->|"1 – 1,000,000"| R1["Shard 1\n(users 1–1M)\nOlder users — less active"]
    Range -->|"1,000,001 – 2,000,000"| R2["Shard 2 ← user 1.5M\n(users 1M–2M)\nModerately active"]
    Range -->|"2,000,001 +"| R3["Shard 3\n(users 2M+)\nNewest users — most active\n⚠️ HOT SPOT"]

    subgraph RangeProps["Range Sharding Characteristics"]
        direction LR
        RP1["✓ Range queries efficient\n(scan single shard)"]
        RP2["✗ Hot spots on new data\n(latest shard takes all writes)"]
        RP3["✓ Simple router logic\n(binary search on ranges)"]
    end

    classDef shard fill:#2980B9,stroke:#1A5276,color:#fff
    classDef active fill:#27AE60,stroke:#1E8449,color:#fff
    classDef hot fill:#E74C3C,stroke:#C0392B,color:#fff

    class R1,R2 shard
    class R3 hot
```

### Directory-Based Sharding

```mermaid
flowchart LR
    Client(["Application"]) --> LookupSvc["Shard Config Service\n(lookup table in DB/cache)"]

    LookupSvc -->|"user 12345 → shard B"| ShardB["Shard B"]
    LookupSvc -->|"user 67890 → shard A"| ShardA["Shard A"]

    subgraph Config["Config Store (Redis / ZooKeeper)"]
        direction TB
        E1["user:1-10000 → shard_a"]
        E2["user:10001-50000 → shard_b"]
        E3["tenant:acme → shard_c"]
    end

    LookupSvc --> Config

    classDef lookup fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef shard fill:#2980B9,stroke:#1A5276,color:#fff
    classDef config fill:#F39C12,stroke:#D68910,color:#fff
    classDef client fill:#1ABC9C,stroke:#148F77,color:#fff

    class LookupSvc lookup
    class ShardA,ShardB shard
    class Config,E1,E2,E3 config
    class Client client
```

**Directory-based pros:** Fully flexible — move individual tenants between shards without resharding everyone. Supports geo-based placement (route EU users to EU shard). **Cons:** Extra network hop for every query; config service is a single point of failure (must be HA + cached).

---

## Query Router: hash(user_id) % N

```mermaid
flowchart TD
    App(["Application Layer"]) --> Router

    subgraph Router["Query Router / Shard Proxy"]
        direction TB
        Parse["Parse request\nextract shard key"]
        Compute["shard_id = hash(user_id) % 3"]
        Parse --> Compute
    end

    Compute -->|"shard_id = 0"| DB0[("Shard 0\nPostgreSQL\n(users 0,3,6,9...)")]
    Compute -->|"shard_id = 1"| DB1[("Shard 1\nPostgreSQL\n(users 1,4,7,10...)")]
    Compute -->|"shard_id = 2"| DB2[("Shard 2\nPostgreSQL\n(users 2,5,8,11...)")]

    DB0 & DB1 & DB2 --> App

    classDef app fill:#F39C12,stroke:#D68910,color:#fff
    classDef router fill:#2980B9,stroke:#1A5276,color:#fff
    classDef db fill:#27AE60,stroke:#1E8449,color:#fff

    class App app
    class Router,Parse,Compute router
    class DB0,DB1,DB2 db
```

---

## Cross-Shard Queries Problem

```mermaid
flowchart TD
    Query["SELECT * FROM orders\nJOIN users ON orders.user_id = users.id\nWHERE orders.total > 1000"] --> Problem

    Problem["Problem: users and orders\nmay be on different shards"]

    Problem --> Option1 & Option2 & Option3

    Option1["Option 1: Application-side join\nFetch from each shard,\njoin in memory\n⚠️ N+1 queries, high latency"]

    Option2["Option 2: Denormalize\nStore user fields in orders table\n(user_name, user_email)\nNo join needed\n⚠️ Data duplication"]

    Option3["Option 3: Co-locate by tenant\nAll user + orders data\non same shard\n✓ All JOINs stay local"]

    classDef problem fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef option fill:#2980B9,stroke:#1A5276,color:#fff
    classDef best fill:#27AE60,stroke:#1E8449,color:#fff

    class Problem problem
    class Option1,Option2 option
    class Option3 best
```

**Cross-shard transactions:** Two-phase commit (2PC) across shards is possible but expensive — adds latency and coordinator failure risk. Better: design the data model to avoid cross-shard transactions. Use the saga pattern (compensating transactions) when cross-shard coordination is unavoidable.

---

## Resharding with Consistent Hashing

```mermaid
flowchart LR
    subgraph Before["Before: 3 shards\nhash(key) % 3"]
        direction TB
        BA["All keys:\nA→S1, B→S2, C→S3\nD→S1, E→S2, F→S3"]
    end

    subgraph NaiveAdd["Naive add: hash(key) % 4\n⚠️ 75% of keys move!"]
        direction TB
        NA["A→S2, B→S3, C→S4\nD→S1, E→S2, F→S3\nMost keys remapped!"]
    end

    subgraph ConsistentAdd["Consistent hashing add node 4\n✓ Only ~25% of keys move"]
        direction TB
        CA["Only keys between S3 and S4\non the ring migrate to S4\n~1/N of total keys"]
    end

    Before -->|"Add shard 4"| NaiveAdd
    Before -->|"Add shard 4"| ConsistentAdd

    classDef before fill:#2980B9,stroke:#1A5276,color:#fff
    classDef bad fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef good fill:#27AE60,stroke:#1E8449,color:#fff

    class Before,BA before
    class NaiveAdd,NA bad
    class ConsistentAdd,CA good
```

**Resharding in practice:**
1. Add new shard to the cluster
2. Copy ~1/N of data from neighbor shard (online, with dual-write)
3. Update routing config to include new shard
4. Drain and validate — old shard serves reads while copy is in progress
5. Switch over — new shard now handles its key range
6. Remove migrated data from old shard

Cassandra uses **vnodes** (each physical node owns 256 virtual ring positions) so resharding is automatic when nodes join/leave.

---

## Hotspot Mitigation

When a specific key naturally receives disproportionate traffic (celebrity tweet, viral post):

| Technique | How | Trade-off |
|---|---|---|
| Key salting | Append `user_id + "_" + random(0,9)` → 10 sub-keys, fan-in on read | Read requires fetching 10 keys and aggregating |
| Virtual shards | Assign hot tenants their own dedicated shard | Requires tenant-aware routing logic |
| CDN / cache layer | Hot data cached before hitting shards | Eventual consistency; cache invalidation complexity |
| Read replicas per shard | Each shard has multiple read replicas | Hot writes still hit one primary |
| Application-level queueing | Queue writes for hot keys, batch flush | Adds write latency for hot users |

**Twitter's approach to celebrity tweets:** Pre-compute and fan-out to follower caches at write time (push model), rather than computing a timeline from all shard queries at read time (pull model). Kylie Jenner's tweet goes to 150 million caches asynchronously, not to 150 million DB reads.

---

## Shard Key Anti-Patterns

| Anti-Pattern | Problem | Better Alternative |
|---|---|---|
| Auto-increment ID | All writes go to the highest shard (monotonic hot spot) | Random UUID, or hash(UUID) |
| Timestamp | Current-time bucket gets 100% of writes | Combine: `hash(user_id) + timestamp bucket` |
| Enum / low-cardinality field | Only N possible shard values (e.g., country has 200 options) | Composite key: `country + user_id` |
| Frequently-updated field | Resharding needed when value changes (e.g., account tier) | Use an immutable field for shard key |
| Globally unique sequential | Still monotonically increasing even if hashed | Twitter Snowflake IDs: time-based but sharded by machine ID |

---

## Real-World Sharding Systems

| System | Sharding Mechanism | Details |
|---|---|---|
| Cassandra | Consistent hashing with vnodes | 256 vnodes per node; automatic rebalancing |
| MongoDB | Chunk-based range sharding | Shards hold chunks; balancer migrates chunks |
| MySQL + Vitess | Application-level hash sharding | Vitess adds query routing, connection pooling |
| DynamoDB | Hash + range key, auto-partitioned | AWS manages shard splits transparently |
| CockroachDB | Range-based, auto-split/rebalance | Raft consensus per range (64 MB chunks) |
| Elasticsearch | Document routing by `_id` | Primary shards fixed at index creation |

---

## Code

The implementation in this module demonstrates hash-based and range-based shard routing in Go:

```
high-level-design/scalability/3.3_db_sharding/
├── router/
│   └── main.go               # Go shard router — hash and range strategies
├── shard/
│   └── main.go               # Go shard service — in-memory key/value store
├── docker-compose.yml        # 1 router + 3 shard services
└── test.sh                   # Write/read tests showing which shard handles each key
```

Run it:

```sh
docker compose up --build -d
bash test.sh
```

---

## Interview Angles

**Q1: How do you handle cross-shard transactions?**
Avoid them by design where possible — co-locate related data on the same shard (e.g., all of a user's data by user_id). When unavoidable, use the saga pattern: a sequence of local transactions with compensating actions on failure. Two-phase commit (2PC) works but adds coordinator complexity and is a single point of failure. Distributed SQL databases (CockroachDB, Spanner) handle this transparently with Raft per shard.

**Q2: What shard key would you choose for a Twitter DM system?**
Shard by `conversation_id` (a hash of the two participants' user IDs, lexicographically sorted). This ensures all messages in a conversation live on the same shard — reads and writes for a conversation hit exactly one shard. Avoid sharding by `sender_id` (half the messages for each conversation would be on different shards, requiring cross-shard joins). Avoid sharding by `timestamp` (hot new-message shard). The conversation_id key is immutable and provides natural co-location.

**Q3: How do you rebalance shards without downtime?**
Use consistent hashing so only `1/N` of keys migrate when adding a shard. Migration process: (1) add new shard, (2) start dual-write to both old and new shard for affected key range, (3) backfill historical data from old to new shard, (4) verify data integrity, (5) cut reads over to new shard, (6) stop dual-write, (7) delete migrated data from old shard. Cassandra and MongoDB's balancer do this automatically. For MySQL + Vitess, use Vitess's online schema change + resharding workflow.

**Q4: What is the hotspot problem and how do you mitigate it?**
A hotspot occurs when a disproportionate fraction of reads or writes routes to a single shard — usually due to a popular entity (viral post, celebrity account) or a bad shard key (timestamp, auto-increment). Mitigations: add a random salt suffix to split the key across N sub-shards (then aggregate on read); add a caching layer in front of the hot shard; use read replicas to scale reads for the hot shard; in extreme cases, give the hot entity its own dedicated shard (virtual tenant isolation).

**Q5: When would you choose range-based over hash-based sharding?**
Choose range-based when your access pattern involves range scans (e.g., "all orders from Jan–Mar", "all events from sensor ID 1000–2000"). Range sharding keeps sequential data local, making range queries efficient (single-shard scan). Choose hash-based for write-heavy workloads where even distribution matters more than range query performance (e.g., user profile data accessed by ID). Many real systems combine both: Cassandra uses hash sharding on the partition key for even distribution, and range ordering within a partition for efficient range scans.

**References**

- [Martin Kleppmann — Designing Data-Intensive Applications, Ch. 6](https://dataintensive.net/)
- [Vitess — Sharding for MySQL at YouTube Scale](https://vitess.io/docs/concepts/sharding/)
- [MongoDB — Sharding Architecture](https://www.mongodb.com/docs/manual/sharding/)
- [Cassandra — Consistent Hashing and vnodes](https://cassandra.apache.org/doc/latest/cassandra/architecture/dynamo.html)
- [Twitter — Snowflake ID Generation](https://blog.twitter.com/engineering/en_us/a/2010/announcing-snowflake)

---

## Architecture Diagrams

> Generated from D2 sources in [`docs/diagrams/src/3.3_db_sharding/`](../../../docs/diagrams/src/3.3_db_sharding/).
> Regenerate: `just diagrams-topic 3.3_db_sharding`.

### Overview

User data split across 4 shards by hash(user_id), each shard with its own replica — plus what makes a good shard key.

![Db Sharding — overview](../../../docs/diagrams/svg/3.3_db_sharding/overview/index.svg)

**`hotspot` scenario:** one celebrity user's shard takes disproportionate traffic — hash distribution balances key count, not load per key.

![Db Sharding — hotspot scenario](../../../docs/diagrams/svg/3.3_db_sharding/overview/hotspot.svg)

**`shard_split` scenario:** a hot shard splits into two, migrating half its key range — the operational answer to growth, painful without consistent hashing.

![Db Sharding — shard_split scenario](../../../docs/diagrams/svg/3.3_db_sharding/overview/shard_split.svg)

### Request Flow

Single-shard queries (one hop, fast) vs scatter-gather queries that fan out to every shard and merge results.

![Db Sharding — request flow](../../../docs/diagrams/svg/3.3_db_sharding/request-flow/index.svg)

**`resharding` scenario:** adding a shard under naive modulo hashing remaps nearly every key; the diagram shows the migration cost that motivates consistent hashing.

![Db Sharding — resharding scenario](../../../docs/diagrams/svg/3.3_db_sharding/request-flow/resharding.svg)
