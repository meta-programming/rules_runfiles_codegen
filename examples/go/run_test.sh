#!/bin/bash
# Run the Go binary and verify it prints the expected output.

# In rules_go, the binary is placed in a subdirectory named <target_name>_/
BINARY="./main_/main"

if [ ! -f "$BINARY" ]; then
  echo "Error: Binary not found at $BINARY"
  exit 1
fi

OUTPUT=$($BINARY)
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
  echo "Error: Binary exited with code $EXIT_CODE"
  exit 1
fi

EXPECTED="Data: dummy content

Helper output: Hello from helper!"
if [ "$OUTPUT" != "$EXPECTED" ]; then
  echo "Error: Unexpected output"
  echo "Expected: $EXPECTED"
  echo "Got:      $OUTPUT"
  exit 1
fi

echo "Quickstart Go example passed!"
