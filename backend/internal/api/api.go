// Package api is the HTTP layer of the calculator. It decodes and validates
// requests, calls the calculator, maps domain errors to HTTP status codes and
// error codes, and encodes JSON.
package api

import (
	"log/slog"
	"net/http"

	"calculator-api/backend/internal/calculator"
)

// MaxBodyBytes is the limit on the request body size.
const MaxBodyBytes = 1 << 20

// Config configures the HTTP handler.
type Config struct {
	// AllowedOrigins is the list of CORS origins.
	AllowedOrigins []string
	// Logger logs each request. Nil means slog.Default().
	Logger *slog.Logger
	// Evaluate calculates an expression. Nil means calculator.Evaluate.
	Evaluate func(string) (calculator.Result, error)
}

// NewHandler returns the full router with its middleware.
func NewHandler(cfg Config) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Evaluate == nil {
		cfg.Evaluate = calculator.Evaluate
	}

	h := &handlers{evaluate: cfg.Evaluate}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/calculate", h.calculate)
	mux.HandleFunc("/api/v1/calculate", methodNotAllowed("POST"))
	mux.HandleFunc("GET /api/v1/operations", h.operations)
	mux.HandleFunc("/api/v1/operations", methodNotAllowed("GET, HEAD"))
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("/healthz", methodNotAllowed("GET, HEAD"))
	mux.HandleFunc("/", notFound)

	return logRequests(cfg.Logger,
		cors(cfg.AllowedOrigins,
			recoverPanics(cfg.Logger,
				limitBody(mux))))
}
