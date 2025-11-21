// Package server provides code for an HTTP server
package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/peter-howell/httpfromtcp/internal/request"
	"github.com/peter-howell/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler Handler
	closed atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Msg string
}

type Handler func(w *response.Writer, req *request.Request)

func (s *Server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	r, err := request.RequestFromReader(conn)

	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := response.NewWriter(conn)
	s.handler(writer, r)
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

func Serve(port uint16, handler Handler) (*Server, error) {
	// Creates a net.Listener and returns a new Server instance. Starts listening for requests inside a goroutine.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler: handler,
	} 
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



