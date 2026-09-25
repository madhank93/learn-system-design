#!/bin/bash

echo "============================================"
echo "  Load Balancer Testing"
echo "============================================"
read -p "Press Enter to continue..."
echo ""

echo "📍 TEST 1: Round Robin (port 8081)"
echo "for i in {1..6}; do curl -s localhost:8081; done"
echo "-------------------------------------------"
for i in {1..6}; do curl -s localhost:8081; done
echo ""
read -p "Press Enter to continue..."
echo ""

echo "📍 TEST 2: Least Connections (port 8082)"
echo "seq 1 100 | xargs -n1 -P20 curl -s localhost:8082"
echo "-------------------------------------------"
seq 1 100 | xargs -n1 -P20 curl -s localhost:8082
echo ""
read -p "Press Enter to continue..."
echo ""

echo "📍 TEST 3: Geo-Based Routing (port 8083)"
echo "curl -s -H 'X-Region: us' localhost:8083"
echo "curl -s -H 'X-Region: eu' localhost:8083"
echo "curl -s -H 'X-Region: jp' localhost:8083"
echo "-------------------------------------------"
echo "US region:"
curl -s -H "X-Region: us" localhost:8083
echo ""
echo "EU region:"
curl -s -H "X-Region: eu" localhost:8083
echo ""
echo "Default region:"
curl -s -H "X-Region: jp" localhost:8083
echo ""
read -p "Press Enter to continue..."
echo ""

echo "📍 TEST 4: Consistent Hashing (port 8084)"
echo "for i in {1..3}; do 
    curl -s --cookie "session_id=alice" localhost:8084
done"
echo "-------------------------------------------"
echo "Alice's requests:"
for i in {1..3}; do 
    curl -s --cookie "session_id=alice" localhost:8084
done
echo ""
echo "Bob's requests:"
for i in {1..3}; do 
    curl -s --cookie "session_id=bob" localhost:8084
done
echo ""
read -p "Press Enter to continue..."
echo ""

echo "============================================"
echo "          All tests completed               "
echo "============================================"
