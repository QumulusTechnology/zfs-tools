package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/QumulusTechnology/zfs-tools/pkg/models"
	"github.com/QumulusTechnology/zfs-tools/pkg/watcher"
)

const (
	// StateFile stores the last timestamp our service was running
	StateFile = "last_run_time.txt"
)

func main() {
	// Parse command line arguments for pools (for simplicity, we'll use defaults if not provided)
	pools := []string{"pool1"}
	if len(os.Args) > 1 {
		pools = os.Args[1:]
	}

	fmt.Printf("Starting continuous ZFS watcher for pools: %v\n", pools)
	fmt.Println("Press Ctrl+C to exit")

	// Get the last run time if available
	lastRunTime := getLastRunTime()
	if !lastRunTime.IsZero() {
		fmt.Printf("Resuming from last run time: %s\n", lastRunTime.Format(time.RFC3339))
	} else {
		fmt.Println("No previous run time found. Starting fresh.")
	}

	// Configure the watcher with last run time to avoid missing events
	cfg := watcher.Config{
		Pools:     pools,
		Interval:  5 * time.Second,
		SinceTime: &lastRunTime,
		// Use the system default zpool path (relies on PATH environment variable)
		ZpoolCmd: watcher.ZpoolCmdDefault,
		// Alternative paths:
		// ZpoolCmd: watcher.ZpoolCmdUsrSbin, // /usr/sbin/zpool
		// ZpoolCmd: watcher.ZpoolCmdSbin,    // /sbin/zpool
		// ZpoolCmd: watcher.ZpoolCmdUsrLocalSbin, // /usr/local/sbin/zpool
		// ZpoolCmd: watcher.ZpoolCommand("/custom/path/to/zpool"), // Custom path
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

		// Save the current time periodically so we can resume from here if the service stops
		saveLastRunTime(time.Now())
	})

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the watcher in a goroutine
	go w.Start()

	// Block until we receive a signal
	<-sigChan
	fmt.Println("\nShutting down...")

	// Save the final runtime before exit
	saveLastRunTime(time.Now())
}

// getLastRunTime retrieves the timestamp of the last successful run
func getLastRunTime() time.Time {
	// Get the directory of the executable
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return time.Time{} // Return zero time if error
	}

	stateFilePath := filepath.Join(dir, StateFile)

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return time.Time{} // Return zero time if file doesn't exist or can't be read
	}

	// Try to parse the timestamp
	unixTime, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return time.Time{} // Return zero time if parsing fails
	}

	return time.Unix(unixTime, 0)
}

// saveLastRunTime saves the current timestamp to a file
func saveLastRunTime(t time.Time) {
	// Get the directory of the executable
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("Error getting directory: %v\n", err)
		return
	}

	stateFilePath := filepath.Join(dir, StateFile)

	// Convert time to Unix timestamp (seconds since epoch)
	unixTime := strconv.FormatInt(t.Unix(), 10)

	// Write to file
	err = os.WriteFile(stateFilePath, []byte(unixTime), 0644)
	if err != nil {
		fmt.Printf("Error saving state: %v\n", err)
	}
}
