#!/usr/bin/env bash
set -e

echo "=== SQL (port 8081) — normalized schema, JOIN queries ==="
echo "List users:"
curl -s http://localhost:8081/users
echo ""
echo "JOIN users + orders:"
curl -s http://localhost:8081/users/orders
echo ""
echo "Insert user:"
curl -s -X POST "http://localhost:8081/users?name=Carol&email=carol@example.com"

echo ""
echo "=== NoSQL (port 8082) — flexible documents, no schema ==="
echo "Store user doc with orders embedded (denormalized):"
curl -s -X PUT http://localhost:8082/doc/u1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","orders":[{"item":"Widget","amount":9.99},{"item":"Gadget","amount":24.99}]}'
curl -s -X PUT http://localhost:8082/doc/u2 \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","email":"bob@example.com","age":30}'

echo ""
echo "Read doc (no JOIN needed — data already together):"
curl -s http://localhost:8082/doc/u1
echo ""
echo "Search by field:"
curl -s "http://localhost:8082/search?field=name&val=Bob"
echo ""
echo "NoSQL allows heterogeneous docs (u2 has 'age', u1 does not):"
curl -s http://localhost:8082/docs
