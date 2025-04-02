# Continuous ZFS Watcher Example

This example demonstrates how to create a continuous ZFS watcher that monitors events in real-time and doesn't exit until the user sends an interrupt signal (Ctrl+C).

## Features

- Monitors ZFS events in real-time
- Detects volume creation, deletion, snapshot creation, deletion, and volume resizing
- Handles multiple pools
- Graceful shutdown with signal handling
- **Resilient to service disruptions** - remembers its state between runs

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
2. It first checks for a previous run timestamp from a state file
3. If a previous run is found, it configures the watcher to catch up from that point
4. It registers a custom event handler to process and display events
5. The watcher runs in a goroutine to monitor events
6. The main thread blocks on a signal channel (SIGINT/SIGTERM)
7. When Ctrl+C is pressed, it gracefully shuts down and saves the current time

## Resilience to Service Disruptions

This example addresses the case where your service might go down and restart. It implements the following resilience mechanisms:

1. **Timestamp Persistence**: The service records the last successful runtime to a file
2. **Event Replay**: On startup, it resumes watching from the last recorded timestamp
3. **No Event Loss**: Events that occurred while the service was down are replayed
4. **Continuous Updates**: The timestamp is updated after each event is processed

This approach ensures no events are missed, making the watcher robust against service interruptions.

## Code Explanation

The example shows:

- How to configure the watcher with a starting time
- How to persist and retrieve the last run timestamp
- How to register custom event handlers
- How to process different event types
- How to handle graceful shutdown with signals
