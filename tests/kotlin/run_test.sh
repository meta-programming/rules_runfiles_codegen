#!/bin/bash
# Fallback to find the executable in runfiles

if [ -f "./integration_test_bin" ]; then
  exec ./integration_test_bin
elif [ -f "./repo/tests/kotlin/integration_test_bin" ]; then
  exec ./repo/tests/kotlin/integration_test_bin
elif [ -f "./integration_test_bin.exe" ]; then
  exec ./integration_test_bin.exe
else
  echo "Cannot find integration_test_bin"
  ls -R
  exit 1
fi
