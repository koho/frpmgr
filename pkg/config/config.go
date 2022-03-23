package config

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
	Complete()
	// GetLogFile returns the log file path of this config.
	GetLogFile() string
	// AutoStart indicates whether this config should be started at boot.
	AutoStart() bool
}

// Section is the interface that must be implemented to build a section in config.
type Section interface {
	// GetName returns the name of this section
	GetName() string
}
