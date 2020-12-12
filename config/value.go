package config

type ServiceState int

const (
	StateUnknown ServiceState = iota
	StateStarted
	StateStopped
	StateStarting
	StateStopping
)

var (
	Version    = "v0.0.0"
	FRPVersion = "v0.0.0"
)
