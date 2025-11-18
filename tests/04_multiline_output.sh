#!/bin/bash
# Test 4: Multiline Output - Test JSON, code blocks, etc.

echo "Starting multiline output test..."
echo ""
echo "=== Test 1: JSON Output ==="
cat <<'EOF'
{
  "status": "processing",
  "data": {
    "id": "12345",
    "name": "Test Item",
    "values": [1, 2, 3, 4, 5],
    "nested": {
      "key1": "value1",
      "key2": "value2"
    }
  }
}
EOF

echo ""
echo "=== Test 2: Stack Trace Simulation ==="
cat <<'EOF' >&2
Error: Something went wrong
    at processData (/app/handler.js:42:15)
    at async handleRequest (/app/server.js:128:9)
    at async Server.handle (/app/server.js:89:5)
EOF

echo ""
echo "=== Test 3: SQL Query ==="
cat <<'EOF'
SELECT
    users.id,
    users.name,
    users.email,
    COUNT(orders.id) as order_count
FROM users
LEFT JOIN orders ON users.id = orders.user_id
WHERE users.created_at > '2024-01-01'
GROUP BY users.id
ORDER BY order_count DESC
LIMIT 10;
EOF

echo ""
echo "=== Test 4: Configuration Dump ==="
cat <<'EOF'
[database]
host = localhost
port = 5432
name = production_db
max_connections = 100

[cache]
enabled = true
ttl = 3600
redis_url = redis://localhost:6379

[logging]
level = info
format = json
output = /var/log/app.log
EOF

echo ""
echo "Test completed successfully!"
exit 0
