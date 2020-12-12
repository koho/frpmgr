package config

type ServiceState int

const (
	StateUnknown ServiceState = iota
	StateStarted
	StateStopped
	StateStarting
	StateStopping
)
