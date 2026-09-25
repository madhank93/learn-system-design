# Indexing

Without an index, every query requires scanning the entire table — O(N) for N rows. An index is a separate data structure that maps column values to row locations, turning lookups into O(log N) and making range queries efficient. The cost: every write (INSERT, UPDATE, DELETE) must update all indexes on the table, and each index consumes additional storage. Choosing which columns to index — and which index structure to use — is one of the most impactful performance decisions in database design.

---

## B-Tree Index

The B-tree (balanced tree) is the dominant index structure in relational databases (PostgreSQL, MySQL InnoDB, SQLite, Oracle). It maintains sorted order, supports both point lookups and range queries, and self-balances to guarantee O(log N) performance.

```mermaid
flowchart TD
    subgraph BTree["B-Tree Index on users.email"]
        direction TB
        Root["Root Node\n[k@example.com | t@example.com]"]

        L1["Internal Node\n[b@ex.com | g@ex.com]"]
        L2["Internal Node\n[m@ex.com | p@ex.com]"]
        L3["Internal Node\n[u@ex.com | z@ex.com]"]

        Leaf1["Leaf: a@ex.com → row 12\nalice@ex.com → row 7"]
        Leaf2["Leaf: b@ex.com → row 3\nfoo@ex.com → row 44"]
        Leaf3["Leaf: g@ex.com → row 91\njane@ex.com → row 18"]
        Leaf4["Leaf: m@ex.com → row 5\nnick@ex.com → row 33"]
        Leaf5["Leaf: p@ex.com → row 67\nroger@ex.com → row 2"]
        Leaf6["Leaf: u@ex.com → row 9\nvera@ex.com → row 41"]

        Root --> L1 & L2 & L3
        L1 --> Leaf1 & Leaf2 & Leaf3
        L2 --> Leaf4 & Leaf5
        L3 --> Leaf6

        LeafLinks["Leaf nodes linked → efficient range scan\ne.g. all emails a-g scanned left-to-right"]
    end

    classDef root fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef internal fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef leaf fill:#27AE60,stroke:#1E8449,color:#fff

    class Root root
    class L1,L2,L3 internal
    class Leaf1,Leaf2,Leaf3,Leaf4,Leaf5,Leaf6 leaf
```

**B-tree properties:**
- Leaf nodes store the actual key → row pointer mappings
- Leaf nodes are linked in sorted order for range scans
- Tree rebalances on every insert/delete to maintain O(log N) height
- Typical branching factor: 100-500 (tree stays shallow even for billions of rows)

---

## LSM-Tree (Log-Structured Merge-Tree)

LSM-trees are optimized for write-heavy workloads. Instead of updating a sorted structure in-place (which requires random I/O), all writes go to an in-memory buffer (MemTable) and are flushed sequentially to disk. Used by: RocksDB, LevelDB, Cassandra, HBase, ScyllaDB, ClickHouse.

```mermaid
flowchart TD
    subgraph Write["Write Path"]
        W["Write\n(key=abc, val=1)"] --> WAL["Write-Ahead Log (WAL)\n(durability — survives crash)"]
        WAL --> Mem["MemTable\n(in-memory sorted tree)\n~64MB threshold"]
        Mem -->|flush when full| L0["SSTable L0\n(immutable sorted file on disk)"]
    end

    subgraph Compact["Compaction (background)"]
        L0 -->|merge + deduplicate| L1["SSTable L1\n(larger, sorted)"]
        L1 -->|merge| L2["SSTable L2\n(even larger)"]
        L2 -->|merge| L3["SSTable L3\n(cold data, largest)"]
    end

    subgraph Read["Read Path"]
        R["Read key=abc"] --> BF["Bloom Filter\n(is key in this SSTable?)"]
        BF -->|probably yes| SIndex["Sparse Index\n(find approximate offset)"]
        SIndex -->|scan block| SST["SSTable\n(binary search in block)"]
        SST -->|found| Result["Return value"]
    end

    classDef write fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef compact fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef read fill:#27AE60,stroke:#1E8449,color:#fff

    class W,WAL,Mem,L0 write
    class L1,L2,L3 compact
    class R,BF,SIndex,SST,Result read
```

---

## B-Tree vs LSM-Tree

| Dimension | B-Tree | LSM-Tree |
|---|---|---|
| Write path | Random I/O (update in-place) | Sequential I/O (append to log) |
| Write amplification | Low (1 write per page) | High (data written multiple times during compaction) |
| Read amplification | Low (O(log N) — single path) | Higher (check MemTable + multiple SSTables) |
| Space amplification | Low | Medium (obsolete data until compaction) |
| Range queries | Excellent (leaf link traversal) | Good (within SSTable; cross-SSTable merge) |
| Point lookups | Excellent | Good with Bloom filters |
| Compaction overhead | None (in-place) | Background CPU/I/O during compaction |
| Best for | Read-heavy, mixed workloads | Write-heavy (time-series, logs, events) |
| Examples | PostgreSQL, MySQL, SQLite | Cassandra, RocksDB, LevelDB, HBase |

---

## Index Types

```mermaid
flowchart LR
    subgraph Types["Index Types"]
        direction TB
        Primary["Primary Index\nOn PK column(s)\nTable physically ordered by PK\n(clustered — InnoDB)\nor separate structure (heap — Postgres)"]
        Secondary["Secondary Index\nOn non-PK column\nStores key → PK pointer\nRequires PK lookup for full row\n(bookmark lookup / index scan + fetch)"]
        Composite["Composite Index\nMultiple columns: (last_name, first_name)\nMatches leftmost prefix queries\n(last_name=X AND first_name=Y) ✓\n(first_name=Y only) ✗"]
        Partial["Partial Index\nIndex rows matching WHERE condition\nCREATE INDEX ON orders(status)\nWHERE status='pending'\nSmaller, faster for filtered queries"]
        Covering["Covering Index (Index-Only Scan)\nIndex includes all columns needed\nby query — no table heap access\nCREATE INDEX ON orders(user_id)\nINCLUDE (amount, created_at)"]
    end

    classDef primary fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef secondary fill:#27AE60,stroke:#1E8449,color:#fff
    classDef composite fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef partial fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef covering fill:#E74C3C,stroke:#C0392B,color:#fff

    class Primary primary
    class Secondary secondary
    class Composite composite
    class Partial partial
    class Covering covering
```

---

## Index Selectivity

**Selectivity** = number of distinct values / total rows. High selectivity = better index.

```mermaid
flowchart LR
    subgraph LowSel["Low Selectivity — Poor Index"]
        direction TB
        LS1["Column: gender\nValues: M / F / other\nDistinct: 3 / 1M rows = 0.0003%"]
        LS2["Index returns 500k rows\nDB may prefer full table scan\n(index scan + 500k heap fetches > seq scan)"]
        LS1 --> LS2
    end

    subgraph HighSel["High Selectivity — Great Index"]
        direction TB
        HS1["Column: email\nValues: unique per user\nDistinct: 1M / 1M rows = 100%"]
        HS2["Index returns 1 row\nO(log N) lookup\n< 1ms on 1B rows"]
        HS1 --> HS2
    end

    classDef bad fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef good fill:#27AE60,stroke:#1E8449,color:#fff

    class LS1,LS2 bad
    class HS1,HS2 good
```

| Column | Selectivity | Index useful? |
|---|---|---|
| email | Very high (unique) | Yes — perfect |
| user_id | Very high (unique) | Yes — primary key |
| status | Low (active/inactive) | Sometimes — partial index |
| created_at | High (many distinct dates) | Yes — for range queries |
| gender | Very low | Rarely — composite with high-selectivity column |
| country | Medium | Depends — useful for country-specific queries |

---

## Full-Text Search — Inverted Index

A B-tree index on a text column supports only prefix matching (`LIKE 'foo%'`), not keyword search (`LIKE '%foo%'`). Full-text search uses an **inverted index**: a mapping from each term to the list of document IDs containing it.

```mermaid
flowchart TD
    subgraph Docs["Documents"]
        D1["doc:1 — 'the quick brown fox'"]
        D2["doc:2 — 'quick brown rabbit'"]
        D3["doc:3 — 'the slow brown turtle'"]
    end

    subgraph Inverted["Inverted Index"]
        direction TB
        T1["'brown' → [1, 2, 3]"]
        T2["'quick' → [1, 2]"]
        T3["'fox' → [1]"]
        T4["'rabbit' → [2]"]
        T5["'slow' → [3]"]
        T6["'turtle' → [3]"]
    end

    Docs --> Inverted

    Query["Query: 'quick brown'"] --> QRes["Intersect: [1,2] ∩ [1,2,3] = [1,2]\nRank by TF-IDF / BM25"]

    classDef doc fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef term fill:#27AE60,stroke:#1E8449,color:#fff
    classDef query fill:#8E44AD,stroke:#6C3483,color:#fff

    class D1,D2,D3 doc
    class T1,T2,T3,T4,T5,T6 term
    class Query,QRes query
```

Used by: Elasticsearch (Lucene), PostgreSQL (`tsvector` + `tsquery`), Solr, Typesense, Meilisearch.

---

## Reading EXPLAIN Output (PostgreSQL)

```sql
-- Without index on email:
EXPLAIN SELECT * FROM users WHERE email = 'alice@example.com';
-- Seq Scan on users  (cost=0.00..18834.00 rows=1 width=64)
--   Filter: (email = 'alice@example.com')

-- After: CREATE INDEX ON users(email);
EXPLAIN SELECT * FROM users WHERE email = 'alice@example.com';
-- Index Scan using users_email_idx on users  (cost=0.43..8.45 rows=1 width=64)
--   Index Cond: (email = 'alice@example.com')
```

| Plan Node | Meaning |
|---|---|
| Seq Scan | Full table scan — O(N) |
| Index Scan | B-tree traversal + heap fetch — O(log N) |
| Index Only Scan | B-tree traversal only (covering index) — O(log N), no heap |
| Bitmap Index Scan | Multiple index entries batch-fetched, then heap accessed in physical order |
| Hash Join / Merge Join | How two tables are joined — prefer over Nested Loop at scale |

---

## Index Overhead and When Not to Index

Every index adds write overhead — each INSERT, UPDATE, DELETE must maintain all indexes on that table. A table with 8 indexes on a write-heavy workload can have 8× write amplification.

**When NOT to index:**
- Small tables (< 1000 rows) — seq scan is often faster
- Columns that are never queried in WHERE, JOIN, or ORDER BY
- Very low-selectivity columns (unless used in composite indexes)
- Write-heavy tables where index maintenance cost outweighs read savings

**Index maintenance strategy:** run `EXPLAIN ANALYZE` on slow queries, identify seq scans on large tables, add targeted indexes. Periodically audit unused indexes (`pg_stat_user_indexes.idx_scan = 0`).

---

## This Module

```
high-level-design/algorithms/7.3_indexing/
├── server/
│   └── main.go    # B-Tree implementation + HTTP API: /insert, /search, /range
├── docker-compose.yml
└── test.sh
```

A B-Tree of order 3 maintains sorted keys across balanced nodes. Point search: O(log n). Range scan: O(log n + k). Used in PostgreSQL, MySQL InnoDB.

```sh
docker compose up --build -d && sleep 1 && bash test.sh
go test -v ./server/
```

**References**

- Martin Kleppmann, "Designing Data-Intensive Applications" — Chapter 3: Storage and Retrieval
- [PostgreSQL Index Types](https://www.postgresql.org/docs/current/indexes-types.html)
- [RocksDB Architecture Guide](https://github.com/facebook/rocksdb/wiki/RocksDB-Architecture-Guide)
- [Elasticsearch: The Definitive Guide — Inverted Index](https://www.elastic.co/guide/en/elasticsearch/guide/current/inverted-index.html)

---

## Interview Angles

**Q1: When would you choose an LSM-tree over a B-tree?**

Choose an LSM-tree for write-heavy workloads: time-series data (IoT sensor readings, metrics), event logs, activity feeds, and message queues. LSM-trees convert random writes into sequential disk I/O (append to log), which is 10-100× faster on HDDs and even significant on SSDs. The trade-off is higher read amplification — reads must check MemTable and potentially multiple SSTable levels, mitigated by Bloom filters. Use B-trees for read-heavy or mixed workloads with frequent point lookups and range queries.

**Q2: What is a covering index and when does it matter?**

A covering index (index-only scan) includes all columns needed by a query within the index itself, so the database never needs to fetch the actual table row. Example: `CREATE INDEX ON orders(user_id) INCLUDE (amount, created_at)` — a query `SELECT amount, created_at FROM orders WHERE user_id = 123` can be answered entirely from the index without touching the table heap. This eliminates random I/O for heap fetches, making reads significantly faster at the cost of larger index size and more write overhead per row.

**Q3: Why does a low-cardinality column make a bad index?**

Low-cardinality columns (gender, status, boolean flags) have few distinct values. An index on `gender = 'M'` returns 50% of the table — the database must then do 500,000 random heap fetches to retrieve the actual rows. The planner often decides a full sequential scan is cheaper (sequential I/O over the whole table beats 500k random I/Os). Exception: use a partial index on a low-cardinality column where one value is rare (`WHERE status='pending'` — most orders are completed, only a small fraction are pending).

**Q4: How does a composite index work and what is the leftmost prefix rule?**

A composite index `(last_name, first_name, email)` is sorted first by `last_name`, then `first_name`, then `email`. It can accelerate queries that use the leftmost prefix of the index: `WHERE last_name = 'Smith'` (uses index), `WHERE last_name = 'Smith' AND first_name = 'John'` (uses index), `WHERE first_name = 'John'` (cannot use index — `last_name` is not specified). Design composite indexes with the most selective column first, and in the order of most common query patterns.

**Q5: How do you find and fix a slow query in production?**

Step 1: identify slow queries from the database slow query log (`pg_stat_statements`, MySQL slow log) — look for high total time or high calls × avg_time. Step 2: run `EXPLAIN ANALYZE` on the query — look for Seq Scan on large tables, high rows estimates vs actual, Nested Loop with large outer results. Step 3: add a targeted index on the filtered/joined column, or rewrite the query to use a more selective predicate. Step 4: verify with `EXPLAIN ANALYZE` post-index that the planner uses Index Scan. Step 5: monitor `pg_stat_user_indexes` to confirm the index is used and measure the latency improvement.

---

## Architecture Diagrams

> Generated from D2 sources in [`docs/diagrams/src/7.3_indexing/`](../../../docs/diagrams/src/7.3_indexing/).
> Regenerate: `just diagrams-topic 7.3_indexing`.

### Overview

B-tree internals — root to internal to leaf in 3–4 levels even for millions of rows — plus when to reach for hash, full-text, or composite indexes.

![Indexing — overview](../../../docs/diagrams/svg/7.3_indexing/overview/index.svg)

**`index_hit` scenario:** the planner walks root → internal → leaf in 3 I/Os — O(log N) with a branching factor of hundreds.

![Indexing — index_hit scenario](../../../docs/diagrams/svg/7.3_indexing/overview/index_hit.svg)

**`index_miss` scenario:** a composite index on (user_id, created_at) is useless for WHERE created_at alone — the leftmost-prefix rule sends the query to a full scan.

![Indexing — index_miss scenario](../../../docs/diagrams/svg/7.3_indexing/overview/index_miss.svg)

### Request Flow

The same query with and without an index: 4ms and 3 page reads vs 8.2s scanning 10M rows.

![Indexing — request flow](../../../docs/diagrams/svg/7.3_indexing/request-flow/index.svg)

**`index_bloat` scenario:** every index slows every write and dead entries accumulate — indexes are paid for at write time, so unused ones are pure cost.

![Indexing — index_bloat scenario](../../../docs/diagrams/svg/7.3_indexing/request-flow/index_bloat.svg)
