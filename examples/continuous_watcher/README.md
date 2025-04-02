# Continuous ZFS Watcher Example

This example demonstrates how to create a continuous ZFS watcher that monitors events in real-time and doesn't exit until the user sends an interrupt signal (Ctrl+C).

## Features

- Monitors ZFS events in real-time
- Detects volume creation, deletion, snapshot creation, deletion, and volume resizing
- Handles multiple pools
- Graceful shutdown with signal handling

## Running the Example

Use the provided run script:

```bash
# Run with default pool (pool1)
./run.sh

# Run with specific pools
./run.sh pool1 pool2 pool3
```

## How It Works

1. The script sets up a ZFS watcher with specified pools
2. It registers a custom event handler to process and display events
3. The watcher runs in a goroutine to monitor events
4. The main thread blocks on a signal channel (SIGINT/SIGTERM)
5. When Ctrl+C is pressed, it gracefully shuts down

## Code Explanation

The example shows:

- How to configure the watcher
- How to register custom event handlers
- How to process different event types
- How to handle graceful shutdown with signals
