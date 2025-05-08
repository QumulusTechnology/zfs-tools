# ZFS Tools

A collection of utilities for working with ZFS pools and datasets.

## ZFS Watcher

The `zfs-watcher` tool monitors ZFS pools for volume and snapshot events by watching the `zpool history` command output. It can detect:

- Volume creation
- Volume deletion
- Snapshot creation
- Snapshot deletion
- Volume resizing

### Installation

```bash
# From the project root
go build -o zfs-watcher ./cmd/zfs-watcher

# Or use the Makefile
make build
```

### Usage

```bash
# Basic usage (monitors pool1 by default)
./zfs-watcher

# Monitor multiple pools
./zfs-watcher --pools pool1,pool2,pool3

# Change check interval (default: 5 seconds)
./zfs-watcher --interval 10

# Output to a file in addition to stdout
./zfs-watcher --output events.log

# Specify a custom zpool command path
./zfs-watcher --zpool-cmd /usr/sbin/zpool

# Show help
./zfs-watcher --help
```

### Examples

```bash
# Monitor two pools with 30-second interval and log to a file
./zfs-watcher --pools pool1,pool2 --interval 30 --output /var/log/zfs-events.log

# Use a specific zpool command path (useful for different OS distributions)
./zfs-watcher --zpool-cmd /usr/local/sbin/zpool

# Using the Makefile
make run ARGS="--pools pool1,pool2 --interval 10 --output events.log"
```

### Running as a Service

A systemd service file is provided in the `config` directory. To install it:

```bash
# Build and install the binary
make build
sudo cp build/zfs-watcher /usr/local/bin/

# Install and enable the service
sudo cp config/zfs-watcher.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable zfs-watcher
sudo systemctl start zfs-watcher

# Check the status
sudo systemctl status zfs-watcher

# View logs
sudo journalctl -u zfs-watcher
```

You may need to modify the service file to match your specific setup, such as adjusting the pools to monitor and the log file location.

## Using as a Library

The ZFS Tools package can be imported and used directly in your Go applications as a library. This allows you to:

1. Monitor ZFS events in real-time with custom handlers
2. Get events from a specific time window (e.g., last 5 minutes)
3. Get events since a specific event occurred

### Importing the Package

```go
import (
    "github.com/QumulusTechnology/zfs-tools/pkg/models"
    "github.com/QumulusTechnology/zfs-tools/pkg/watcher"
)
```

### Real-time Monitoring

```go
// Configure the watcher
cfg := watcher.Config{
    Pools:    []string{"pool1", "pool2"},
    Interval: 5 * time.Second,
    ZpoolCmd: watcher.ZpoolCmdDefault, // Use the default path (relies on system PATH)
    // Alternatives:
    // ZpoolCmd: watcher.ZpoolCmdUsrSbin, // Use /usr/sbin/zpool
    // ZpoolCmd: watcher.ZpoolCmdSbin,    // Use /sbin/zpool
    // ZpoolCmd: watcher.ZpoolCmdUsrLocalSbin, // Use /usr/local/sbin/zpool
    // ZpoolCmd: watcher.ZpoolCommand("/custom/path/to/zpool"), // Custom path
}

// Create the watcher
w := watcher.New(cfg)

// Add a custom event handler
w.AddEventHandler(func(event models.ZFSEvent) {
    // Do something with each event
    if event.Type == models.EventVolumeCreated {
        // Handle volume creation...
    }
})

// Start watching in a goroutine
go w.Start()
```

### Getting Recent Events

```go
// Get events from the last 5 minutes
events, err := w.GetRecentEvents(5 * time.Minute)
if err != nil {
    log.Fatalf("Failed to get recent events: %v", err)
}

// Process the events
for _, event := range events {
    fmt.Printf("[%s] %s: %s\n",
        event.Timestamp.Format("2006-01-02 15:04:05"),
        event.Type,
        event.Target)
}
```

### Getting Events Since a Specific Time

```go
// Get events since yesterday
sinceTime := time.Now().AddDate(0, 0, -1)
events, err := w.GetEventsSince(sinceTime)
if err != nil {
    log.Fatalf("Failed to get events: %v", err)
}
```

### Getting Events Since a Specific Event

```go
// Get events since a specific command was executed
events, err := w.GetEventsSinceEvent("zfs create -s -V 5244048KB pool1/volume-xyz")
if err != nil {
    log.Fatalf("Failed to get events: %v", err)
}
```

### Complete Example

See the [examples directory](./examples) for complete examples of using the package as a library.

## Project Structure

This project is structured as a monorepo with the following organization:

```
zfs-tools/
├── cmd/                   # Command-line tools
│   └── zfs-watcher/       # The ZFS watcher CLI
├── pkg/                   # Shared packages
│   ├── models/            # Data models
│   └── watcher/           # ZFS event watching implementation
├── examples/              # Library usage examples
│   └── library_usage.go   # Example code for using as a library
├── config/                # Configuration files
│   └── zfs-watcher.service # Systemd service file
├── build/                 # Build output (generated)
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── README.md              # Project documentation
```

## License

MIT
