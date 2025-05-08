package watcher

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/QumulusTechnology/zfs-tools/pkg/models"
)

// Config holds the watcher configuration
type Config struct {
	// Pools to monitor
	Pools []string

	// Interval between checks
	Interval time.Duration

	// SinceTime if set, only report events since this time
	SinceTime *time.Time

	// SinceEvent if set, only report events since this event command was seen
	SinceEvent string

	// ZpoolCmd specifies the path to the zpool command
	ZpoolCmd ZpoolCommand
}

// EventHandler is a function that handles ZFS events
type EventHandler func(event models.ZFSEvent)

// Watcher watches for ZFS changes
type Watcher struct {
	config          Config
	lastEvents      map[string]time.Time
	eventHandlers   []EventHandler
	volumeCreateRE  *regexp.Regexp
	volumeDestroyRE *regexp.Regexp
	snapshotRE      *regexp.Regexp
	volResizeRE     *regexp.Regexp
	seenSinceEvent  bool
}

// New creates a new ZFS watcher
func New(config Config) *Watcher {
	// Set default zpool command if not specified
	if config.ZpoolCmd == "" {
		config.ZpoolCmd = ZpoolCmdDefault
	}

	return &Watcher{
		config:     config,
		lastEvents: make(map[string]time.Time),
		// Detect volume creation
		volumeCreateRE: regexp.MustCompile(`zfs create\s+.*?(-s -V\s+(\d+)KB.*?)?pool\d+\/(volume-[a-f0-9\-]+_\d+)`),

		// Detect volume deletion
		volumeDestroyRE: regexp.MustCompile(`zfs destroy\s+pool\d+\/(volume-[a-f0-9\-]+_\d+)(?:\@(snapshot-[a-f0-9\-]+))?`),

		// Detect snapshot creation
		snapshotRE: regexp.MustCompile(`zfs snapshot\s+pool\d+\/(volume-[a-f0-9\-]+_\d+)\@(snapshot-[a-f0-9\-]+)`),

		// Detect volume resize
		volResizeRE: regexp.MustCompile(`zfs set volsize=(\d+)KB\s+pool\d+\/(volume-[a-f0-9\-]+_\d+)`),

		// Set to true if no sinceEvent is specified
		seenSinceEvent: config.SinceEvent == "",
	}
}

// AddEventHandler adds a new event handler
func (w *Watcher) AddEventHandler(handler EventHandler) {
	w.eventHandlers = append(w.eventHandlers, handler)
}

// Start begins monitoring ZFS changes
func (w *Watcher) Start() {
	log.Printf("Starting ZFS watcher for pools: %v", w.config.Pools)
	log.Printf("Monitoring for volume and snapshot events")

	// Initialize with current history (gather initial state, don't report)
	for _, pool := range w.config.Pools {
		w.processPoolHistory(pool, true)
	}

	// Start periodic checking
	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	for range ticker.C {
		for _, pool := range w.config.Pools {
			w.processPoolHistory(pool, false)
		}
	}
}

// GetEventsSince returns events since the specified time
func (w *Watcher) GetEventsSince(sinceTime time.Time) ([]models.ZFSEvent, error) {
	var events []models.ZFSEvent

	for _, pool := range w.config.Pools {
		poolEvents, err := w.getPoolEventsSince(pool, sinceTime)
		if err != nil {
			return nil, err
		}
		events = append(events, poolEvents...)
	}

	return events, nil
}

// GetEventsSinceEvent returns events since the specified event command
func (w *Watcher) GetEventsSinceEvent(sinceEventCmd string) ([]models.ZFSEvent, error) {
	var events []models.ZFSEvent
	var foundEvent bool

	for _, pool := range w.config.Pools {
		cmd := exec.Command(string(w.config.ZpoolCmd), "history", pool)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("error getting history for pool %s: %v", pool, err)
		}

		var poolEvents []models.ZFSEvent
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "History for") || line == "" {
				continue
			}

			// Check if this is the event we're looking for
			if !foundEvent && strings.Contains(line, sinceEventCmd) {
				foundEvent = true
				continue // Skip the marker event itself
			}

			// Only collect events after the marker
			if foundEvent {
				event, err := w.parseEvent(line, pool)
				if err == nil {
					poolEvents = append(poolEvents, event)
				}
			}
		}

		events = append(events, poolEvents...)
	}

	if !foundEvent {
		return nil, fmt.Errorf("event '%s' not found in history", sinceEventCmd)
	}

	return events, nil
}

// getPoolEventsSince returns events for a specific pool since the given time
func (w *Watcher) getPoolEventsSince(pool string, sinceTime time.Time) ([]models.ZFSEvent, error) {
	var events []models.ZFSEvent

	cmd := exec.Command(string(w.config.ZpoolCmd), "history", pool)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting history for pool %s: %v", pool, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "History for") || line == "" {
			continue
		}

		event, err := w.parseEvent(line, pool)
		if err != nil {
			continue
		}

		// Only include events after sinceTime
		if event.Timestamp.After(sinceTime) {
			events = append(events, event)
		}
	}

	return events, nil
}

// GetRecentEvents returns events from the last duration
func (w *Watcher) GetRecentEvents(duration time.Duration) ([]models.ZFSEvent, error) {
	sinceTime := time.Now().Add(-duration)
	return w.GetEventsSince(sinceTime)
}

// processPoolHistory processes the history of a ZFS pool
func (w *Watcher) processPoolHistory(pool string, initialize bool) {
	cmd := exec.Command(string(w.config.ZpoolCmd), "history", pool)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting history for pool %s: %v", pool, err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "History for") || line == "" {
			continue
		}

		// Check if this is the sinceEvent if we're looking for one
		if !w.seenSinceEvent && w.config.SinceEvent != "" && strings.Contains(line, w.config.SinceEvent) {
			w.seenSinceEvent = true
			continue // Skip the marker event itself
		}

		event, err := w.parseEvent(line, pool)
		if err != nil {
			continue
		}

		// Skip if we've seen this event before
		lastTime, seen := w.lastEvents[event.Command]
		if seen && !event.Timestamp.After(lastTime) {
			continue
		}

		// Skip if before sinceTime
		if w.config.SinceTime != nil && !event.Timestamp.After(*w.config.SinceTime) {
			continue
		}

		// Skip if we haven't seen the sinceEvent yet
		if !w.seenSinceEvent {
			continue
		}

		// Update last seen time
		w.lastEvents[event.Command] = event.Timestamp

		// Skip reporting events during initialization
		if initialize {
			continue
		}

		// Notify handlers
		for _, handler := range w.eventHandlers {
			handler(event)
		}
	}
}

// parseEvent parses a line from zpool history output
func (w *Watcher) parseEvent(line string, pool string) (models.ZFSEvent, error) {
	event := models.ZFSEvent{Pool: pool}

	// Parse timestamp and command
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return event, fmt.Errorf("invalid format")
	}

	timestamp := parts[0]
	command := parts[1]

	// Parse the timestamp
	t, err := time.Parse("2006-01-02.15:04:05", timestamp)
	if err != nil {
		return event, fmt.Errorf("invalid timestamp: %v", err)
	}
	event.Timestamp = t
	event.Command = command

	// Try to match different types of events

	// Check for volume creation
	if match := w.volumeCreateRE.FindStringSubmatch(command); match != nil {
		event.Type = models.EventVolumeCreated
		event.VolumeID = match[3]
		event.Target = event.VolumeID
		if match[2] != "" {
			event.Size = match[2]
		}
		return event, nil
	}

	// Check for volume resize
	if match := w.volResizeRE.FindStringSubmatch(command); match != nil {
		event.Type = models.EventVolumeResized
		event.VolumeID = match[2]
		event.Target = event.VolumeID
		event.Size = match[1]
		return event, nil
	}

	// Check for snapshot creation
	if match := w.snapshotRE.FindStringSubmatch(command); match != nil {
		event.Type = models.EventSnapshotCreated
		event.VolumeID = match[1]
		event.SnapshotID = match[2]
		event.Target = fmt.Sprintf("%s@%s", event.VolumeID, event.SnapshotID)
		return event, nil
	}

	// Check for volume/snapshot deletion
	if match := w.volumeDestroyRE.FindStringSubmatch(command); match != nil {
		event.VolumeID = match[1]
		if match[2] != "" {
			// Snapshot deletion
			event.Type = models.EventSnapshotDeleted
			event.SnapshotID = match[2]
			event.Target = fmt.Sprintf("%s@%s", event.VolumeID, event.SnapshotID)
		} else {
			// Volume deletion
			event.Type = models.EventVolumeDeleted
			event.Target = event.VolumeID
		}
		return event, nil
	}

	return event, fmt.Errorf("not a matching event")
}

// LoggingHandler returns an event handler that logs events
func LoggingHandler() EventHandler {
	return func(event models.ZFSEvent) {
		timeStr := event.Timestamp.Format("2006-01-02 15:04:05")

		switch event.Type {
		case models.EventVolumeCreated:
			log.Printf("[%s] Volume created: %s on pool %s", timeStr, event.Target, event.Pool)
		case models.EventVolumeDeleted:
			log.Printf("[%s] Volume deleted: %s on pool %s", timeStr, event.Target, event.Pool)
		case models.EventSnapshotCreated:
			log.Printf("[%s] Snapshot created: %s on pool %s", timeStr, event.Target, event.Pool)
		case models.EventSnapshotDeleted:
			log.Printf("[%s] Snapshot deleted: %s on pool %s", timeStr, event.Target, event.Pool)
		case models.EventVolumeResized:
			log.Printf("[%s] Volume resized: %s to %sKB on pool %s", timeStr, event.Target, event.Size, event.Pool)
		}
	}
}
