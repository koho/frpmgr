package config

import "gopkg.in/ini.v1"

func init() {
	ini.PrettyFormat = false
	ini.PrettyEqual = true
}

// Config is the interface that a config must implement to support management.
type Config interface {
	// Items returns all sections in this config. which must be a slice of pointer to struct.
	Items() interface{}
	// ItemAt returns the section in this config for the given index.
	ItemAt(index int) interface{}
	// DeleteItem deletes the section for the given index.
	DeleteItem(index int)
	// AddItem adds a section to this config.
	AddItem(item interface{}) bool
	// Save serializes this config and saves to the given path.
	Save(path string) error
	// Complete prunes and completes this config.
	// When "read" is true, the config should be completed for a file loaded from source.
	// Otherwise, it should be completed for file written to disk.
	Complete(read bool)
	// GetLogFile returns the log file path of this config.
	GetLogFile() string
	// AutoStart indicates whether this config should be started at boot.
	AutoStart() bool
	// GetExpiry returns the expiry days of this config.
	GetExpiry() uint
	// Copy creates a new copy of this config.
	Copy(all bool) Config
}
