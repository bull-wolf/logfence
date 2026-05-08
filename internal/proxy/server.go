package proxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Config holds the proxy server configuration.
type Config struct {
	ListenAddr string
	UpstreamURL string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Server wraps an HTTP server with the redacting proxy handler.
type Server struct {
	httpServer *http.Server
}

// NewServer constructs a Server that forwards redacted traffic to the upstream.
func NewServer(cfg Config, handler *Handler) (*Server, error) {
	upstream, err := url.Parse(cfg.UpstreamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL %q: %w", cfg.UpstreamURL, err)
	}

	rp := httputil.NewSingleHostReverseProxy(upstream)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[logfence] upstream error: %v", err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}

	// Wire: request → redact handler → reverse proxy
	handler.next = rp

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{httpServer: srv}, nil
}

// Start begins listening and blocks until the server stops.
func (s *Server) Start() error {
	log.Printf("[logfence] proxy listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
