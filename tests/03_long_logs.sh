#!/bin/bash
# Test 3: Long Logs - Generate many log lines

echo "Starting long log test..."
echo "Generating 100 log entries..."

for i in {1..100}; do
    if [ $((i % 10)) -eq 0 ]; then
        echo "Progress: $i/100 entries generated"
    fi
    echo "Log entry #$i: Processing item with ID $(uuidgen | cut -d'-' -f1)"

    # Add occasional warnings
    if [ $((i % 25)) -eq 0 ]; then
        echo "Warning: Checkpoint reached at entry $i" >&2
    fi

    # Small delay every 20 entries
    if [ $((i % 20)) -eq 0 ]; then
        sleep 0.1
    fi
done

echo "All 100 entries processed successfully"
echo "Test completed!"
exit 0
