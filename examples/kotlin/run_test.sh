#!/bin/bash
# Run the Kotlin binary and verify output.
# In rules_kotlin, the binary wrapper script is placed directly in the package directory.
BINARY="./main"

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
Helper output: Hello from helper!
FileSet paths: [dummy.txt, info.txt]
FileSet dummy content: dummy content"
if [ "$OUTPUT" != "$EXPECTED" ]; then
  echo "Error: Unexpected output"
  echo "Expected: $EXPECTED"
  echo "Got:      $OUTPUT"
  exit 1
fi

echo "Quickstart Kotlin example passed!"
