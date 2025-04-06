package ipc

import (
	"encoding/gob"
	"net"

	"github.com/Microsoft/go-winio"
	"github.com/fatedier/frp/client"
)

type Server struct {
	listener net.Listener
	exporter client.StatusExporter
}

func NewServer(name string, exporter client.StatusExporter) (*Server, error) {
	listener, err := winio.ListenPipe(`\\.\pipe\`+name, &winio.PipeConfig{
		MessageMode:      true,
		InputBufferSize:  1024,
		OutputBufferSize: 1024,
	})
	if err != nil {
		return nil, err
	}
	return &Server{listener, exporter}, nil
}

func (s *Server) Run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	for {
		var names []string
		if err := gob.NewDecoder(conn).Decode(&names); err != nil {
			return
		}
		msg := make([]ProxyMessage, 0, len(names))
		for _, name := range names {
			if status, _ := s.exporter.GetProxyStatus(name); status != nil {
				msg = append(msg, ProxyMessage{
					Name:       status.Name,
					Type:       status.Type,
					Status:     status.Phase,
					Err:        status.Err,
					RemoteAddr: status.RemoteAddr,
				})
			}
		}
		if err := gob.NewEncoder(conn).Encode(msg); err != nil {
			return
		}
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}
