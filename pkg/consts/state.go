package consts

// ConfigState is the state of FRP daemon service
type ConfigState int

const (
	StateUnknown ConfigState = iota
	StateStarted
	StateStopped
	StateStarting
	StateStopping
)
