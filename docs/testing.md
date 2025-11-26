---
layout: default
title: Testing
nav_order: 9
description: "Testing guide for SwiftLog - running and writing tests"
---

# SwiftLog Testing Guide

This guide covers testing SwiftLog and writing tests for new features.

## Test Suite Overview

SwiftLog includes a comprehensive integration test suite in the `tests/` directory to validate logging capabilities across different scenarios.

## Prerequisites

### 1. Build CLI Tool

```bash
cd cli
go build -o swiftlog
```

### 2. Start Backend Services

```bash
cd /path/to/swiftlog
make start
```

### 3. Configure CLI

```bash
./cli/swiftlog config set --token test-token --server localhost:50051
```

---

## Running Tests

### Run All Tests

```bash
cd tests
./run_all_tests.sh
```

This will run all test scripts sequentially and report results.

### Run Individual Test

```bash
cd /path/to/swiftlog
./cli/swiftlog run --project test-project --group 01_simple_test -- bash tests/01_simple_test.sh
```

---

## Test Scripts

### 01_simple_test.sh

**Purpose:** Basic success case with simple stdout messages

**Tests:**
- Sequential output
- Sleep delays
- Success exit code (0)

**Expected Results:**
- Clean stdout logs
- Exit code 0
- All messages captured in order

**Run:**
```bash
./cli/swiftlog run --project test --group simple -- bash tests/01_simple_test.sh
```

---

### 02_stderr_test.sh

**Purpose:** Mixed stdout and stderr output with error exit

**Tests:**
- Warning messages to stderr
- Error messages to stderr
- Mixed stream handling
- Error exit code (1)

**Expected Results:**
- Both stdout and stderr logged separately
- Proper stream labeling
- Exit code 1 preserved

**Run:**
```bash
./cli/swiftlog run --project test --group stderr -- bash tests/02_stderr_test.sh
```

---

### 03_long_logs.sh

**Purpose:** Volume test with high number of log entries

**Tests:**
- 100+ log entries
- Progress markers
- Periodic warnings
- Performance under load

**Expected Results:**
- All 100 entries logged successfully
- No dropped messages
- Warning messages captured
- Acceptable performance

**Run:**
```bash
./cli/swiftlog run --project test --group long -- bash tests/03_long_logs.sh
```

---

### 04_multiline_output.sh

**Purpose:** Complex multiline output formats

**Tests:**
- JSON objects
- Stack traces
- SQL queries
- Configuration files
- YAML/INI formats

**Expected Results:**
- Multiline content preserved correctly
- No line breaks lost
- Proper formatting maintained

**Run:**
```bash
./cli/swiftlog run --project test --group multiline -- bash tests/04_multiline_output.sh
```

---

## Test Output

Each test creates:
- A project named `test-project` (auto-created if not exists)
- A group named after the test script
- A run record with all captured logs
- Metadata including exit code and timestamps

### Viewing Results

**Via Frontend:**
1. Open http://localhost:3000
2. Navigate to Projects → test-project
3. Select the group (e.g., `01_simple_test`)
4. View the run logs

**Via API:**
```bash
# List runs in a group
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/groups/{group_id}/runs

# Get logs for a run
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/runs/{run_id}/logs
```

---

## Writing New Tests

### Create a Test Script

1. **Create file:** `tests/05_your_test.sh`

```bash
#!/bin/bash
set -e

echo "Starting your test..."

# Your test logic here
echo "Testing feature X"
sleep 1
echo "Feature X works!"

echo "Test completed successfully"
exit 0
```

2. **Make executable:**
```bash
chmod +x tests/05_your_test.sh
```

3. **Add to test runner:**
Edit `tests/run_all_tests.sh` to include your test.

### Test Script Guidelines

**Good Practices:**
- Use descriptive names: `05_websocket_streaming.sh`
- Include comments explaining what's being tested
- Use `set -e` to fail on errors
- Output clear messages for each step
- Use appropriate exit codes

**Exit Codes:**
- `0` - Success
- `1` - General failure
- `2+` - Specific error codes

**Output:**
- Use `echo` for stdout messages
- Use `echo "error" >&2` for stderr messages
- Include timing with `sleep` for async tests

### Example Test Template

```bash
#!/bin/bash
# Test: <Description of what this tests>
# Expected: <Expected behavior>

set -e

echo "=== Test: Your Feature ==="

# Setup
echo "Setting up test environment..."
# Your setup code

# Test execution
echo "Running test..."
# Your test code

# Verification
echo "Verifying results..."
# Your verification code

# Cleanup
echo "Cleaning up..."
# Your cleanup code

echo "=== Test completed successfully ==="
exit 0
```

---

## Unit Tests

### Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/auth/...

# Verbose output
go test -v ./...

# Run with race detector
go test -race ./...
```

### Frontend Tests

```bash
cd frontend

# Run tests
npm test

# Run with coverage
npm test -- --coverage

# Run in watch mode
npm test -- --watch

# Type checking
npm run type-check

# Linting
npm run lint
```

### CLI Tests

```bash
cd cli

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

---

## Continuous Integration

SwiftLog uses GitHub Actions for CI. Tests run automatically on:
- Pull requests
- Pushes to main branch
- Version tag pushes

### Running CI Locally

Use [act](https://github.com/nektos/act) to test workflows locally:

```bash
# Install act
brew install act  # macOS
# or see: https://github.com/nektos/act

# Run CI workflow
act push

# Run specific job
act -j test
```

---

## Test Coverage

### Backend Coverage

```bash
cd backend

# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Coverage summary
go tool cover -func=coverage.out
```

### Frontend Coverage

```bash
cd frontend

# Generate coverage report
npm test -- --coverage

# View in browser
open coverage/lcov-report/index.html
```

---

## Performance Testing

### CLI Overhead Test

Test CLI performance overhead:

```bash
# Without SwiftLog
time ./your-script.sh

# With SwiftLog
time ./cli/swiftlog run --project test --group perf -- ./your-script.sh
```

**Target:** <5% overhead for scripts running >10 seconds

### Load Testing

Test concurrent connections:

```bash
# Run multiple CLI clients simultaneously
for i in {1..100}; do
  ./cli/swiftlog run --project load-test --group "client-$i" -- \
    bash -c "echo 'Client $i'; sleep 5" &
done
wait
```

### API Load Testing

Use tools like `ab` or `wrk`:

```bash
# Apache Bench
ab -n 1000 -c 10 -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/projects

# wrk
wrk -t4 -c100 -d30s -H "Authorization: Bearer TOKEN" \
  http://localhost:8080/api/v1/projects
```

---

## Debugging Tests

### Enable Debug Logging

```bash
# Backend services
export LOG_LEVEL=debug
docker compose restart

# CLI
./cli/swiftlog run --debug --project test -- your-script.sh
```

### View Service Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f ingestor

# Last 100 lines
docker compose logs --tail=100 api
```

### Database Inspection

```bash
# Connect to PostgreSQL
docker compose exec postgres psql -U swiftlog -d swiftlog

# Check test data
SELECT * FROM projects WHERE name = 'test-project';
SELECT * FROM log_groups WHERE project_id = 'project-uuid';
SELECT * FROM log_runs WHERE group_id = 'group-uuid';
```

### Loki Inspection

```bash
# Query logs directly from Loki
curl -G -s "http://localhost:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={project="test-project"}' \
  --data-urlencode "start=$(date -u -d '1 hour ago' +%s)000000000" \
  --data-urlencode "end=$(date -u +%s)000000000"
```

---

## Troubleshooting Tests

### Tests Not Running

**Check CLI is built:**
```bash
./cli/swiftlog --version
```

**Check services are running:**
```bash
docker compose ps
```

**Check configuration:**
```bash
./cli/swiftlog config get
```

### Logs Not Appearing

**Check Ingestor logs:**
```bash
docker compose logs ingestor
```

**Verify Loki is healthy:**
```bash
curl http://localhost:3100/ready
```

**Check database connection:**
```bash
docker compose exec postgres psql -U swiftlog -d swiftlog -c '\dt'
```

### Authentication Errors

**Verify token exists:**
```bash
docker compose exec postgres psql -U swiftlog -d swiftlog -c \
  "SELECT * FROM api_tokens WHERE token_hash = encode(sha256('test-token'::bytea), 'hex');"
```

**Create test token:**
```bash
docker compose exec postgres psql -U swiftlog -d swiftlog -c \
  "INSERT INTO api_tokens (user_id, token_hash, name)
   SELECT id, encode(sha256('test-token'::bytea), 'hex'), 'Test Token'
   FROM users LIMIT 1;"
```

---

## Best Practices

### For Integration Tests

1. **Isolate tests** - Each test should be independent
2. **Clean up** - Remove test data after tests (optional)
3. **Use unique names** - Avoid conflicts between test runs
4. **Test edge cases** - Empty input, large input, special characters
5. **Document expected behavior** - Clear comments and descriptions

### For Unit Tests

1. **Test one thing** - Each test should validate one behavior
2. **Use descriptive names** - Test names should explain what's tested
3. **Mock external dependencies** - Database, APIs, etc.
4. **Test error cases** - Not just happy paths
5. **Keep tests fast** - Unit tests should run in milliseconds

### For Performance Tests

1. **Establish baselines** - Know your starting point
2. **Test realistic scenarios** - Use real-world data sizes
3. **Monitor resources** - CPU, memory, network
4. **Test at scale** - Simulate production load
5. **Automate testing** - Run regularly to catch regressions

---

## Related Documentation

- [Getting Started](getting-started)
- [Development Guide](development)
- [Architecture](architecture)
- [CLI Guide](cli-guide)