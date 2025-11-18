#!/bin/bash
# Test Runner - Execute all SwiftLog test scripts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="test-project"
TESTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_PATH="${TESTS_DIR}/../cli/swiftlog"

# Check if CLI exists
if [ ! -f "$CLI_PATH" ]; then
    echo -e "${RED}Error: SwiftLog CLI not found at $CLI_PATH${NC}"
    echo "Please build the CLI first: cd cli && go build -o swiftlog"
    exit 1
fi

# Make all test scripts executable
echo -e "${BLUE}Making test scripts executable...${NC}"
chmod +x "${TESTS_DIR}"/*.sh

# Test results tracking
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a single test
run_test() {
    local test_script="$1"
    local test_name=$(basename "$test_script" .sh)
    local group_name="$test_name"

    echo ""
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}Running: $test_name${NC}"
    echo -e "${BLUE}======================================${NC}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Run the test through SwiftLog CLI
    if "$CLI_PATH" run --project "$PROJECT_NAME" --group "$group_name" -- bash "$test_script"; then
        echo -e "${GREEN}✓ Test passed: $test_name${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        local exit_code=$?
        # Some tests are expected to fail (like stderr_test)
        if [ "$test_name" = "02_stderr_test" ] && [ $exit_code -eq 1 ]; then
            echo -e "${GREEN}✓ Test passed (expected failure): $test_name${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            echo -e "${RED}✗ Test failed: $test_name (exit code: $exit_code)${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    fi
}

# Main execution
echo -e "${YELLOW}Starting SwiftLog Test Suite${NC}"
echo -e "${YELLOW}Project: $PROJECT_NAME${NC}"
echo ""

# Find and run all test scripts
for test_script in "${TESTS_DIR}"/[0-9][0-9]_*.sh; do
    if [ -f "$test_script" ]; then
        run_test "$test_script"
    fi
done

# Print summary
echo ""
echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}======================================${NC}"
echo -e "Total Tests:  $TOTAL_TESTS"
echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
else
    echo -e "Failed:       $FAILED_TESTS"
fi
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed ✗${NC}"
    exit 1
fi
