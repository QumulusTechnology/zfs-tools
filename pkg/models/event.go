package models

import (
	"time"
)

// EventType represents the type of ZFS event
type EventType string

const (
	// EventVolumeCreated represents a volume creation event
	EventVolumeCreated EventType = "VOLUME_CREATED"

	// EventVolumeDeleted represents a volume deletion event
	EventVolumeDeleted EventType = "VOLUME_DELETED"

	// EventSnapshotCreated represents a snapshot creation event
	EventSnapshotCreated EventType = "SNAPSHOT_CREATED"

	// EventSnapshotDeleted represents a snapshot deletion event
	EventSnapshotDeleted EventType = "SNAPSHOT_DELETED"

	// EventVolumeResized represents a volume resize event
	EventVolumeResized EventType = "VOLUME_RESIZED"
)

// ZFSEvent represents a parsed ZFS event
type ZFSEvent struct {
	// Timestamp is when the event occurred
	Timestamp time.Time

	// Command is the raw ZFS command executed
	Command string

	// Pool is the ZFS pool name
	Pool string

	// Type is the event type
	Type EventType

	// Target is the volume or snapshot ID
	Target string

	// VolumeID is the volume identifier (without snapshot suffix if present)
	VolumeID string

	// SnapshotID is the snapshot identifier (if applicable)
	SnapshotID string

	// Size is the size in KB for resize events (if applicable)
	Size string
}
