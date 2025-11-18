# SwiftLog CLI Documentation

The SwiftLog CLI is a command-line tool for streaming script logs to the SwiftLog platform in real-time.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Commands](#commands)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Installation

### Build from Source

```bash
# From the project root
cd cli
go build -o swiftlog

# Or use the Makefile
cd ..
make cli
```

### Install Globally

```bash
# Copy to system path
sudo cp swiftlog /usr/local/bin/

# Verify installation
swiftlog --version
```

### Binary Releases

Download pre-built binaries from the [Releases page](https://github.com/your-repo/swiftlog/releases).

## Configuration

The CLI requires configuration to connect to the SwiftLog backend.

### Initial Setup

1. **Obtain an API Token**

   You need to create an API token first. You can do this by:

   - Using the SwiftLog web interface (recommended)
   - Directly inserting into the database (for testing)
   - Using the API endpoint (if you already have a token)

   **For testing, create a token via database:**
   ```bash
   docker compose exec postgres psql -U swiftlog -d swiftlog -c \
     "INSERT INTO api_tokens (user_id, token_hash, name)
      SELECT id, encode(sha256('test-token'::bytea), 'hex'), 'CLI Test Token'
      FROM users LIMIT 1
      RETURNING id;"
   ```

2. **Configure the CLI**

   ```bash
   swiftlog config set --token YOUR_API_TOKEN --server localhost:50051
   ```

### Configuration File

The configuration is stored at:
- **Linux/macOS**: `~/.swiftlog/config.yaml`
- **Windows**: `%USERPROFILE%\.swiftlog\config.yaml`

**Config file format:**
```yaml
server: localhost:50051
token: YOUR_API_TOKEN
```

### View Configuration

```bash
# Show current configuration
swiftlog config get

# Show config file path
swiftlog config path
```

## Usage

### Basic Syntax

```bash
swiftlog run [flags] -- <command> [args...]
```

The `--` separator is required to distinguish CLI flags from the command you want to run.

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--project` | `-p` | Project name | `default` |
| `--group` | `-g` | Log group name | `default` |
| `--server` | `-s` | gRPC server address | From config |
| `--token` | `-t` | API token | From config |

## Commands

### run

Execute a command and stream its logs to SwiftLog.

```bash
swiftlog run --project <project> --group <group> -- <command>
```

**Behavior:**
- Captures both stdout and stderr
- Streams logs in real-time to the backend
- Preserves the original command's exit code
- Displays output to your terminal (passthrough)

### config

Manage CLI configuration.

**Subcommands:**

```bash
# Set configuration
swiftlog config set --token <token> --server <server>

# Get current configuration
swiftlog config get

# Show config file path
swiftlog config path
```

### version

Display CLI version information.

```bash
swiftlog version
```

### help

Display help information.

```bash
swiftlog help
swiftlog run --help
swiftlog config --help
```

## Examples

### Basic Script Execution

```bash
# Run a simple command
swiftlog run --project myapp --group tests -- echo "Hello, SwiftLog!"

# Run a shell script
swiftlog run --project webapp --group build -- ./build.sh

# Run with default project and group
swiftlog run -- npm test
```

### Long-Running Processes

```bash
# Run a Python training script
swiftlog run --project ml --group training -- python train_model.py --epochs 100

# Run a data processing job
swiftlog run --project etl --group daily -- python process_data.py
```

### Complex Commands

```bash
# Command with pipes (use bash -c)
swiftlog run --project data --group analysis -- bash -c "cat data.csv | grep ERROR | wc -l"

# Command with environment variables
swiftlog run --project api --group deploy -- bash -c "ENV=prod ./deploy.sh"

# Command with redirections
swiftlog run --project backup -- bash -c "mysqldump mydb > backup.sql 2>&1"
```

### Continuous Integration

```bash
# In a CI/CD pipeline
swiftlog run --project myapp --group ci-build -- make build
swiftlog run --project myapp --group ci-test -- make test
swiftlog run --project myapp --group ci-deploy -- ./deploy.sh
```

### Error Handling

```bash
# Script that may fail
swiftlog run --project api --group healthcheck -- curl https://api.example.com/health

# The CLI preserves exit codes
if swiftlog run --project myapp --group tests -- pytest; then
    echo "Tests passed!"
else
    echo "Tests failed!"
    exit 1
fi
```

## Advanced Usage

### Using Different Servers

```bash
# Override server for a specific run
swiftlog run --server production.example.com:50051 --project prod -- ./deploy.sh

# Use different environments
swiftlog run --server staging:50051 --project staging --group deploy -- ./deploy.sh
```

### Organizing Logs

**Recommended hierarchy:**

```
Project → Group → Runs

Examples:
- Project: "webapp"
  - Group: "frontend-build"
  - Group: "backend-build"
  - Group: "e2e-tests"

- Project: "ml-pipeline"
  - Group: "data-preprocessing"
  - Group: "model-training"
  - Group: "model-evaluation"

- Project: "scheduled-jobs"
  - Group: "daily-backup"
  - Group: "weekly-report"
  - Group: "monthly-cleanup"
```

### Testing with Test Suite

```bash
# Run the included test scripts
cd tests
./run_all_tests.sh

# Or run individual tests
cd /path/to/swiftlog
./cli/swiftlog run --project test-project --group 01_simple_test -- bash tests/01_simple_test.sh
./cli/swiftlog run --project test-project --group 02_stderr_test -- bash tests/02_stderr_test.sh
./cli/swiftlog run --project test-project --group 03_long_logs -- bash tests/03_long_logs.sh
./cli/swiftlog run --project test-project --group 04_multiline_output -- bash tests/04_multiline_output.sh
```

## Troubleshooting

### Connection Refused

**Problem:** `Failed to connect to server: connection refused`

**Solutions:**
1. Check if the Ingestor service is running:
   ```bash
   docker compose ps ingestor
   ```

2. Verify server address in config:
   ```bash
   swiftlog config get
   ```

3. Test connectivity:
   ```bash
   telnet localhost 50051
   # or
   nc -zv localhost 50051
   ```

### Authentication Failed

**Problem:** `Authentication failed: invalid token`

**Solutions:**
1. Verify your token is correct:
   ```bash
   swiftlog config get
   ```

2. Check if token exists in database:
   ```bash
   docker compose exec postgres psql -U swiftlog -d swiftlog -c \
     "SELECT name, created_at FROM api_tokens WHERE token_hash = encode(sha256('YOUR_TOKEN'::bytea), 'hex');"
   ```

3. Create a new token and update config:
   ```bash
   swiftlog config set --token NEW_TOKEN --server localhost:50051
   ```

### Logs Not Appearing

**Problem:** Command runs but logs don't appear in the web interface

**Solutions:**
1. Check Ingestor logs:
   ```bash
   docker compose logs ingestor
   ```

2. Verify Loki is running:
   ```bash
   docker compose ps loki
   ```

3. Check for errors in the CLI output

4. Verify project and group exist:
   ```bash
   curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/v1/projects
   ```

### File Already Closed Error

**Problem:** `Error reading output: read |0: file already closed`

**Status:** This is a benign error that occurs when pipes close naturally at command completion. The CLI filters these errors automatically. If you see this error, it can be safely ignored - your logs are still captured correctly.

### Slow Performance

**Problem:** CLI adds noticeable overhead to script execution

**Causes:**
- Very high-frequency output (thousands of lines per second)
- Network latency to SwiftLog server
- Server resource constraints

**Solutions:**
1. Check network latency:
   ```bash
   ping <server-address>
   ```

2. Verify server resources:
   ```bash
   docker compose ps
   docker stats
   ```

3. For very high-volume logs, consider batching (future feature)

### Config File Issues

**Problem:** Can't find or read config file

**Solutions:**
1. Check config file location:
   ```bash
   swiftlog config path
   ```

2. Create config directory manually:
   ```bash
   mkdir -p ~/.swiftlog
   ```

3. Set configuration again:
   ```bash
   swiftlog config set --token YOUR_TOKEN --server localhost:50051
   ```

## Integration Examples

### Cron Jobs

```bash
# Add to crontab
0 2 * * * /usr/local/bin/swiftlog run --project backups --group daily -- /opt/scripts/backup.sh

# Or use a wrapper script
#!/bin/bash
export PATH=/usr/local/bin:$PATH
swiftlog run --project scheduled --group $(basename $0) -- "$@"
```

### GitHub Actions

```yaml
name: Test and Deploy

on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Install SwiftLog CLI
        run: |
          curl -L https://github.com/your-repo/swiftlog/releases/latest/download/swiftlog-linux-amd64 -o swiftlog
          chmod +x swiftlog
          sudo mv swiftlog /usr/local/bin/

      - name: Configure SwiftLog
        run: swiftlog config set --token ${{ secrets.SWIFTLOG_TOKEN }} --server logs.example.com:50051

      - name: Run tests
        run: swiftlog run --project my-app --group github-ci -- npm test
```

### Docker Container

```dockerfile
FROM golang:1.21-alpine

# Install SwiftLog CLI
COPY --from=swiftlog-cli:latest /swiftlog /usr/local/bin/swiftlog

# Configure
RUN swiftlog config set --token ${SWIFTLOG_TOKEN} --server ${SWIFTLOG_SERVER}

# Run your application with SwiftLog
CMD swiftlog run --project myapp --group docker -- ./myapp
```

## Technical Details

### Architecture

```
┌─────────────────────┐
│   Your Command      │
│   (subprocess)      │
└──────────┬──────────┘
           │
           │ stdout/stderr
           │
┌──────────▼──────────┐
│   SwiftLog CLI      │
│                     │
│  - Pipes output     │
│  - Buffers lines    │
│  - Streams via gRPC │
└──────────┬──────────┘
           │
           │ gRPC Stream
           │
┌──────────▼──────────┐
│  Ingestor Service   │
│  (Backend)          │
└─────────────────────┘
```

### Output Handling

- **Buffering**: Line-buffered (flushes on newline)
- **Encoding**: UTF-8
- **Labeling**: Stdout → `[STDOUT]`, Stderr → `[STDERR]`
- **Exit Code**: Preserved from the wrapped command

### gRPC Protocol

The CLI uses a gRPC streaming connection defined in `proto/ingestor.proto`:

```protobuf
service LogIngestor {
  rpc StreamLogs(stream LogRequest) returns (LogResponse);
}
```

## Building from Source

### Prerequisites

- Go 1.21 or higher
- Protocol Buffers compiler (protoc)

### Build Steps

```bash
# Clone repository
git clone <repository-url>
cd swiftlog/cli

# Download dependencies
go mod download

# Generate protobuf code (if needed)
protoc --go_out=. --go-grpc_out=. proto/ingestor.proto

# Build
go build -o swiftlog

# Run tests
go test ./...
```

### Cross-Compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o swiftlog-linux-amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -o swiftlog-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o swiftlog-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o swiftlog-windows-amd64.exe
```

## Support

- **Documentation**: [Main README](../README.md)
- **API Documentation**: [docs/API.md](../docs/API.md)
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions

## License

See [LICENSE](../LICENSE) in the project root.
