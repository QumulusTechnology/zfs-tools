package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/QumulusTechnology/zfs-tools/pkg/models"
	"github.com/QumulusTechnology/zfs-tools/pkg/watcher"
	"github.com/spf13/cobra"
)

var (
	pools          []string
	interval       int
	outputFile     string
	outputToFile   bool
	outputToStdout bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "zfs-watcher",
		Short: "Monitor ZFS events by watching pool history",
		Long: `A utility that monitors ZFS events by watching the zpool history command.
This tool will detect volume creation, snapshot creation, volume deletion, 
snapshot deletion, and volume resize events.`,
		Run: run,
	}

	rootCmd.Flags().StringSliceVarP(&pools, "pools", "p", []string{"pool1"}, "ZFS pools to monitor (comma-separated)")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 5, "Check interval in seconds")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().BoolVarP(&outputToStdout, "stdout", "s", true, "Output to stdout")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	// Configure output
	outputToFile = outputFile != ""

	if outputToFile {
		f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	// Configure the watcher
	cfg := watcher.Config{
		Pools:    pools,
		Interval: time.Duration(interval) * time.Second,
	}

	// Create and set up watcher
	w := watcher.New(cfg)

	// Add the default logging handler
	w.AddEventHandler(watcher.LoggingHandler())

	// Add our custom handler for file output if needed
	if outputToFile {
		w.AddEventHandler(fileOutputHandler(outputFile))
	}

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the watcher in a goroutine
	go w.Start()

	fmt.Printf("ZFS watcher started. Monitoring pools: %s\n", strings.Join(pools, ", "))
	fmt.Println("Press Ctrl+C to exit.")

	// Wait for SIGINT or SIGTERM
	<-sigChan
	fmt.Println("\nShutting down...")
}

// fileOutputHandler returns an event handler that writes events to a file
func fileOutputHandler(filepath string) watcher.EventHandler {
	return func(event models.ZFSEvent) {
		timeStr := event.Timestamp.Format("2006-01-02 15:04:05")
		var line string

		switch event.Type {
		case models.EventVolumeCreated:
			line = fmt.Sprintf("[%s] Volume created: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventVolumeDeleted:
			line = fmt.Sprintf("[%s] Volume deleted: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventSnapshotCreated:
			line = fmt.Sprintf("[%s] Snapshot created: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventSnapshotDeleted:
			line = fmt.Sprintf("[%s] Snapshot deleted: %s on pool %s\n", timeStr, event.Target, event.Pool)
		case models.EventVolumeResized:
			line = fmt.Sprintf("[%s] Volume resized: %s to %sKB on pool %s\n", timeStr, event.Target, event.Size, event.Pool)
		}

		// Append the line to the file
		f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()

		if _, err := f.WriteString(line); err != nil {
			return
		}
	}
}
