package main

import (
	"fmt"
	"log"
	"time"

	"github.com/QumulusTechnology/zfs-tools/pkg/models"
	"github.com/QumulusTechnology/zfs-tools/pkg/watcher"
)

func main() {
	// Example 1: Real-time monitoring with custom handler
	exampleRealTimeMonitoring()

	// Example 2: Getting events from the last 5 minutes
	exampleGetRecentEvents()

	// Example 3: Getting events since a specific time
	exampleGetEventsSinceTime()

	// Example 4: Getting events since a specific event
	exampleGetEventsSinceEvent()

	// Example 5: Real-time monitoring but only since a specific time
	exampleMonitoringSinceTime()
}

// exampleRealTimeMonitoring demonstrates how to watch for ZFS events in real-time
func exampleRealTimeMonitoring() {
	fmt.Println("\n=== Example 1: Real-time Monitoring with Custom Handler ===")

	// Configure the watcher
	cfg := watcher.Config{
		Pools:    []string{"pool1"},
		Interval: 5 * time.Second,
	}

	// Create the watcher
	w := watcher.New(cfg)

	// Add a custom event handler
	w.AddEventHandler(func(event models.ZFSEvent) {
		// You can process events here as they happen
		// For example, send to a message queue, update a database, etc.
		switch event.Type {
		case models.EventVolumeCreated:
			fmt.Printf("CUSTOM HANDLER: Volume %s was created\n", event.VolumeID)
			// Your custom logic here...
		case models.EventVolumeDeleted:
			fmt.Printf("CUSTOM HANDLER: Volume %s was deleted\n", event.VolumeID)
		// Your custom logic here...
		case models.EventSnapshotCreated:
			fmt.Printf("CUSTOM HANDLER: Snapshot %s was created\n", event.SnapshotID)
			// Your custom logic here...
		}
	})

	// Add the standard logging handler
	w.AddEventHandler(watcher.LoggingHandler())

	// Start the watcher in a goroutine
	go w.Start()

	fmt.Println("Watching for ZFS events in real-time...")
	fmt.Println("This would normally run indefinitely, but for the example we'll sleep for a few seconds.")

	// Sleep for a few seconds to simulate running (would normally run indefinitely)
	time.Sleep(3 * time.Second)
}

// exampleGetRecentEvents demonstrates how to get events from the last few minutes
func exampleGetRecentEvents() {
	fmt.Println("\n=== Example 2: Getting Recent Events (last 5 minutes) ===")

	// Configure the watcher
	cfg := watcher.Config{
		Pools: []string{"pool1"},
	}

	// Create the watcher
	w := watcher.New(cfg)

	// Get events from the last 5 minutes
	events, err := w.GetRecentEvents(5 * time.Minute)
	if err != nil {
		log.Fatalf("Failed to get recent events: %v", err)
	}

	fmt.Printf("Found %d events in the last 5 minutes:\n", len(events))
	for i, event := range events {
		fmt.Printf("%d. [%s] %s (%s)\n",
			i+1,
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Type,
			event.Target)
	}
}

// exampleGetEventsSinceTime demonstrates how to get events since a specific time
func exampleGetEventsSinceTime() {
	fmt.Println("\n=== Example 3: Getting Events Since a Specific Time ===")

	// Configure the watcher
	cfg := watcher.Config{
		Pools: []string{"pool1"},
	}

	// Create the watcher
	w := watcher.New(cfg)

	// Get events since yesterday
	sinceTime := time.Now().AddDate(0, 0, -1) // 1 day ago
	events, err := w.GetEventsSince(sinceTime)
	if err != nil {
		log.Fatalf("Failed to get events since time: %v", err)
	}

	fmt.Printf("Found %d events since %s:\n",
		len(events),
		sinceTime.Format("2006-01-02 15:04:05"))

	for i, event := range events {
		fmt.Printf("%d. [%s] %s (%s)\n",
			i+1,
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Type,
			event.Target)
	}
}

// exampleGetEventsSinceEvent demonstrates how to get events since a specific event
func exampleGetEventsSinceEvent() {
	fmt.Println("\n=== Example 4: Getting Events Since a Specific Event ===")

	// Configure the watcher
	cfg := watcher.Config{
		Pools: []string{"pool1"},
	}

	// Create the watcher
	w := watcher.New(cfg)

	// Get events since a specific event command
	// This would typically be a specific ZFS command you're interested in
	sinceEventCmd := "zfs create -s -V 5244048KB"
	events, err := w.GetEventsSinceEvent(sinceEventCmd)
	if err != nil {
		fmt.Printf("Failed to get events since event: %v\n", err)
		return
	}

	fmt.Printf("Found %d events since command '%s':\n", len(events), sinceEventCmd)
	for i, event := range events {
		fmt.Printf("%d. [%s] %s (%s)\n",
			i+1,
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.Type,
			event.Target)
	}
}

// exampleMonitoringSinceTime demonstrates real-time monitoring starting from a specific time
func exampleMonitoringSinceTime() {
	fmt.Println("\n=== Example 5: Real-time Monitoring Starting From a Specific Time ===")

	// Set a time to start monitoring from (1 hour ago)
	sinceTime := time.Now().Add(-1 * time.Hour)

	// Configure the watcher with a specific start time
	cfg := watcher.Config{
		Pools:     []string{"pool1"},
		Interval:  5 * time.Second,
		SinceTime: &sinceTime,
	}

	// Create the watcher
	w := watcher.New(cfg)

	// Add the standard logging handler
	w.AddEventHandler(watcher.LoggingHandler())

	fmt.Printf("Watching for ZFS events since %s...\n",
		sinceTime.Format("2006-01-02 15:04:05"))
	fmt.Println("This would normally run indefinitely, but for the example we'll just sleep for a few seconds.")

	// Start the watcher in a goroutine
	go w.Start()

	// Sleep for a few seconds to simulate running (would normally run indefinitely)
	time.Sleep(3 * time.Second)
}
