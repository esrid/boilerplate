package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"yourapp/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HTTPServer implements the Server interface
type HTTPServer struct {
	db     *pgxpool.Pool
	logger *slog.Logger
	cfg    *config.AppConfig

	router Router
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(db *pgxpool.Pool, logger *slog.Logger, config *config.AppConfig) *HTTPServer {
	router := NewRouter(db, config, logger)

	return &HTTPServer{
		db:     db,
		logger: logger,
		cfg:    config,
		router: router,
	}
}

// Start starts the HTTP server and blocks until shutdown
func (s *HTTPServer) Start() error {
	s.server = &http.Server{
		Addr:         ":" + s.cfg.Port,
		Handler:      s.router.Handler(),
		IdleTimeout:  s.cfg.HTTP.IdleTimeout,
		ReadTimeout:  s.cfg.HTTP.ReadTimeout,
		WriteTimeout: s.cfg.HTTP.WriteTimeout,
	}

	// Channel for server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("starting server",
			slog.String("port", s.cfg.Port),
		)

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	return s.waitForShutdown(serverErr)
}

// waitForShutdown waits for interrupt signal or server error and handles graceful shutdown
func (s *HTTPServer) waitForShutdown(serverErr <-chan error) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		s.logger.Error("server error", slog.String("error", err.Error()))
		return err
	case sig := <-quit:
		s.logger.Info("received shutdown signal", slog.String("signal", sig.String()))
		return s.gracefulShutdown()
	}
}

// gracefulShutdown handles graceful server shutdown
func (s *HTTPServer) gracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	s.logger.Info("shutting down server...")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		return err
	}

	// Close database connection if it exists
	if s.db != nil {
		s.db.Close()
	}

	s.logger.Info("server shutdown completed")
	return nil
}
