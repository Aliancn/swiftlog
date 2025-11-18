#!/bin/bash
# Test 2: Stderr Testing - Mix of stdout and stderr

echo "Starting stderr test..."
echo "Normal output to stdout"
echo "Warning: This is a warning message" >&2
sleep 0.5
echo "Continuing normal processing..."
echo "Error: Something went wrong here" >&2
sleep 0.5
echo "Attempting recovery..."
echo "Fatal error: Critical failure detected" >&2
echo "Test completed with errors"
exit 1
