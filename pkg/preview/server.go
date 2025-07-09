package preview

import (
	"context"
	"fmt"
	"gogallery/pkg/config"
	"gogallery/pkg/datastore"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Server struct {
	*mux.Router
	*datastore.DataStore
	server   *http.Server
	listener net.Listener
	mu       sync.Mutex
	running  bool
	addr     string // actual address (host:port)
}

// generateAddr generates a random address for the server to listen on.
func generateAddr() string {
	host := "127.0.0.1"
	if config.Config.UI.Public {
		host = "0.0.0.0"
	}
	// Always return host:0 so OS picks a free port
	return fmt.Sprintf("%s:0", host)
}

func NewServer(db *datastore.DataStore) *Server {
	router := mux.NewRouter()
	addr := generateAddr()
	server := &Server{
		Router:    router,
		DataStore: db,
		addr:      addr,
		mu:        sync.Mutex{},
	}
	server.Setup()
	return server
}

// Start starts the server on a random port. If public is true, binds to 0.0.0.0, else localhost.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil // already running
	}
	// Listen on the requested address to get a free port, then use that for http.Server
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln
	s.addr = ln.Addr().String()
	s.server = &http.Server{Handler: s.Router}
	s.running = true
	go func() {
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()
	return nil
}

// Stop stops the server if running.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	err := s.server.Shutdown(context.Background())
	s.listener.Close()
	s.running = false
	return err
}

// Addr returns the actual address the server is listening on.
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

func (s *Server) Status() (running bool, addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running, s.addr
}
