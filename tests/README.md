# SwiftLog Test Suite

This directory contains test scripts to validate SwiftLog's logging capabilities across different scenarios.

## Test Scripts

### 01_simple_test.sh
Basic success case with simple stdout messages.
- Tests: Sequential output, sleep delays, success exit code
- Expected: Clean stdout logs, exit code 0

### 02_stderr_test.sh
Mixed stdout and stderr output with error exit.
- Tests: Warning messages, error messages, mixed stream handling
- Expected: Both stdout and stderr logged separately, exit code 1

### 03_long_logs.sh
Volume test with 100 log entries.
- Tests: High volume logging, progress markers, periodic warnings
- Expected: All 100 entries logged, warning messages captured

### 04_multiline_output.sh
Complex multiline output formats.
- Tests: JSON objects, stack traces, SQL queries, configuration files
- Expected: Multiline content preserved correctly

## Running Tests

### Run All Tests
```bash
cd /Users/aliancn/code/swiftlog/tests
./run_all_tests.sh
```

### Run Individual Test
```bash
cd /Users/aliancn/code/swiftlog
./cli/swiftlog run --project test-project --group my-test -- bash tests/01_simple_test.sh
```

## Prerequisites

1. Build the SwiftLog CLI:
```bash
cd cli
go build -o swiftlog
```

2. Ensure backend services are running:
```bash
cd /Users/aliancn/code/swiftlog
make start
```

## Test Output

Each test creates:
- A project named `test-project` (automatically created if not exists)
- A group named after the test (e.g., `01_simple_test`)
- A run record with all captured logs
- AI analysis can be triggered from the frontend

## Viewing Results

After running tests, view results in:
- Frontend: http://localhost:3000
- Navigate to: Projects → test-project → Select group → View runs

## Expected Behavior

- **01_simple_test**: Should complete successfully with exit code 0
- **02_stderr_test**: Should complete with exit code 1 (expected)
- **03_long_logs**: Should capture all 100 log entries
- **04_multiline_output**: Should preserve multiline formatting

## Adding New Tests

1. Create a new script: `05_your_test.sh`
2. Add execute permissions: `chmod +x 05_your_test.sh`
3. Follow the naming convention: `[number]_[description].sh`
4. The test runner will automatically discover and run it
