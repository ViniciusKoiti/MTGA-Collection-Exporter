// Package httpapi owns the central HTTP server lifecycle (OpenSpec
// add-central-go-platform, task 3.1): header, request, idle, shutdown,
// body, decompression and concurrency limits are explicit budgets, and
// shutdown drains in-flight requests before Run returns.
package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

// Config declares every server budget; incomplete budgets are refused.
type Config struct {
	Addr          string
	ReadHeader    time.Duration
	Read          time.Duration
	Write         time.Duration
	Idle          time.Duration
	Shutdown      time.Duration
	MaxBodyBytes  int64
	MaxConcurrent int64
}

func (c Config) validate() error {
	if c.Addr == "" || c.ReadHeader <= 0 || c.Read <= 0 || c.Write <= 0 ||
		c.Idle <= 0 || c.Shutdown <= 0 || c.MaxBodyBytes <= 0 ||
		c.MaxConcurrent < 1 {
		return fmt.Errorf("httpapi: incomplete server budgets: %+v", c)
	}
	return nil
}

// Server wraps net/http with explicit budgets and a drained shutdown.
type Server struct {
	cfg  Config
	http *http.Server
	ln   net.Listener
}

// New wires the handler behind the limit middleware and binds the
// listener up front so the port is owned before Run starts serving.
func New(cfg Config, handler http.Handler) (*Server, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("httpapi: listen %s: %w", cfg.Addr, err)
	}
	srv := &http.Server{
		Handler:           withLimits(cfg, handler),
		ReadHeaderTimeout: cfg.ReadHeader,
		ReadTimeout:       cfg.Read,
		WriteTimeout:      cfg.Write,
		IdleTimeout:       cfg.Idle,
	}
	return &Server{cfg: cfg, http: srv, ln: ln}, nil
}

// Addr reports the bound address (supports ":0" in tests).
func (s *Server) Addr() string { return s.ln.Addr().String() }

// Run serves until ctx is cancelled, then shuts down within the
// shutdown budget, draining in-flight requests before returning.
func (s *Server) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := s.http.Serve(s.ln); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-gctx.Done()
		stop, cancel := context.WithTimeout(context.Background(), s.cfg.Shutdown)
		defer cancel()
		return s.http.Shutdown(stop)
	})
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
