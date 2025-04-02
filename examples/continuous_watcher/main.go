package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/QumulusTechnology/zfs-tools/pkg/models"
	"github.com/QumulusTechnology/zfs-tools/pkg/watcher"
)

func main() {
	// Parse command line arguments for pools (for simplicity, we'll use defaults if not provided)
	pools := []string{"pool1"}
	if len(os.Args) > 1 {
		pools = os.Args[1:]
	}

	fmt.Printf("Starting continuous ZFS watcher for pools: %v\n", pools)
	fmt.Println("Press Ctrl+C to exit")

	// Configure the watcher
	cfg := watcher.Config{
		Pools:    pools,
		Interval: 5 * time.Second,
	}

	// Create the watcher
	w := watcher.New(cfg)

	// Add a custom event handler
	w.AddEventHandler(func(event models.ZFSEvent) {
		timeStr := event.Timestamp.Format("2006-01-02 15:04:05")

		switch event.Type {
		case models.EventVolumeCreated:
			fmt.Printf("[%s] Volume created: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventVolumeDeleted:
			fmt.Printf("[%s] Volume deleted: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventSnapshotCreated:
			fmt.Printf("[%s] Snapshot created: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventSnapshotDeleted:
			fmt.Printf("[%s] Snapshot deleted: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventVolumeResized:
			fmt.Printf("[%s] Volume resized: %s to %sKB on pool %s\n", timeStr, event.Target, event.Size, event.Pool)
		default:
			fmt.Printf("[%s] Unknown event: %s\n", timeStr, event.Command)
		}
	})

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the watcher in a goroutine
	go w.Start()

	// Block until we receive a signal
	<-sigChan
	fmt.Println("\nShutting down...")
}
