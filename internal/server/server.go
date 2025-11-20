// Package server provides code for an HTTP server
package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/peter-howell/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed atomic.Bool
}

func (s *Server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()
	h := response.GetDefaultHeaders(0)

	_ = response.WriteStatusLine(conn, response.StatusOK)

	_ = response.WriteHeaders(conn, h)
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func Serve(port uint16) (*Server, error) {
	// Creates a net.Listener and returns a new Server instance. Starts listening for requests inside a goroutine.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{listener: listener} 
	go server.listen()
	return server, nil
}


func (s *Server) Close() error {
	s.closed.Store(true) 
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}



