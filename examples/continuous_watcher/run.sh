#!/bin/bash

# Go to the project root directory
cd "$(dirname "$0")/../.."

# Build the example
echo "Building continuous watcher example..."
go build -o build/continuous_watcher examples/continuous_watcher

# Run the example with any provided pool names
echo "Running continuous watcher..."
if [ $# -eq 0 ]; then
  # No arguments provided, use default pool
  ./build/continuous_watcher
else
  # Pass all script arguments to the watcher
  ./build/continuous_watcher "$@"
fi 