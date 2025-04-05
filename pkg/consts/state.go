package consts

// ConfigState is the state of FRP daemon service
type ConfigState int

const (
	ConfigStateUnknown ConfigState = iota
	ConfigStateStarted
	ConfigStateStopped
	ConfigStateStarting
	ConfigStateStopping
)
