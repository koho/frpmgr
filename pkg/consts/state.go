package consts

// ServiceState is the state of FRP daemon service
type ServiceState int

const (
	StateUnknown ServiceState = iota
	StateStarted
	StateStopped
	StateStarting
	StateStopping
)
