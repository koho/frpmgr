package ipc

import "context"

// ProxyMessage is the status information of a proxy.
type ProxyMessage struct {
	Name       string
	Type       string
	Status     string
	Err        string
	RemoteAddr string
}

// Client is used to query proxy state from the frp client.
// It may be a pipe client or HTTP client.
type Client interface {
	// SetCallback changes the callback function for the response message.
	SetCallback(cb func([]ProxyMessage))
	// Run the client in blocking mode.
	Run(ctx context.Context)
	// Probe triggers a query request immediately.
	Probe(ctx context.Context)
}
