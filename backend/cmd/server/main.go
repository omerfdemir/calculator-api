// Command server runs the calculator HTTP API.
//
// Usage:
//
//	server              serve the API and /healthz on $PORT
//	server healthcheck  call /healthz on $PORT and exit with 0 or 1
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"calculator-api/backend/internal/api"
)

const shutdownTimeout = 10 * time.Second

type config struct {
	port           string
	allowedOrigins []string
}

func loadConfig(getenv func(string) string) config {
	cfg := config{port: getenv("PORT"), allowedOrigins: []string{"http://localhost:5173"}}
	if cfg.port == "" {
		cfg.port = "8080"
	}
	if origins := getenv("ALLOWED_ORIGINS"); origins != "" {
		cfg.allowedOrigins = nil
		for _, origin := range strings.Split(origins, ",") {
			if origin = strings.TrimSpace(origin); origin != "" {
				cfg.allowedOrigins = append(cfg.allowedOrigins, origin)
			}
		}
	}
	return cfg
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer) error {
	cfg := loadConfig(getenv)
	if len(args) > 0 {
		if args[0] == "healthcheck" {
			return healthcheck(ctx, cfg.port)
		}
		return fmt.Errorf("unknown command %q", args[0])
	}
	return serve(ctx, cfg, slog.New(slog.NewJSONHandler(stdout, nil)))
}

// serve runs the server until ctx ends or a SIGINT or SIGTERM arrives, and
// then shuts down gracefully.
func serve(ctx context.Context, cfg config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", ":"+cfg.port)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	server := &http.Server{
		Handler:           api.NewHandler(api.Config{AllowedOrigins: cfg.allowedOrigins, Logger: logger}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	logger.Info("server started", slog.String("addr", listener.Addr().String()))

	select {
	case err := <-served:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	logger.Info("server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-served; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

// healthcheck calls /healthz. The distroless image has no shell or curl, so
// the Docker health check uses this subcommand.
func healthcheck(ctx context.Context, port string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port+"/healthz", http.NoBody)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: status %d", resp.StatusCode)
	}
	return nil
}
