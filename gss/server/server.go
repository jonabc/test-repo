package server

import (
	"net/http"
)

// Server defines a standard HTTP server
type Server struct {
	handler http.Handler
	addr    string
}

// NewHTTPServer creates a new HTTP server with the given parameters. Addr expects a string with a colon prefixing a port number.
func NewHTTPServer(handler http.Handler, addr string) *Server {
	server := Server{
		handler: handler,
		addr:    addr,
	}
	return &server
}

// Run registers the handlers and starts the server.
func (s *Server) Run() error {
	// set up HTTP server
	return http.ListenAndServe(s.addr, s.handler)
}
