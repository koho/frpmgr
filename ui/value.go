package ui

type ServiceState int

const (
	StateUnknown ServiceState = iota
	StateStarted
	StateStopped
	StateStarting
	StateStopping
)
