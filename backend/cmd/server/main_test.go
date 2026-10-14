package main

import (
	"bytes"
	"context"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

// freePort returns a TCP port that is free at the time of the call.
func freePort(t *testing.T) string {
	t.Helper()
	l, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// syncBuffer is a bytes.Buffer that is safe for the server goroutine and the test.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestServerServesHealthAndShutsDownGracefully(t *testing.T) {
	port := freePort(t)
	getenv := env(map[string]string{"PORT": port})
	var logs syncBuffer

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- run(ctx, nil, getenv, &logs) }()

	// The healthcheck subcommand succeeds when the server is up.
	deadline := time.Now().Add(5 * time.Second)
	for {
		err := run(t.Context(), []string{"healthcheck"}, getenv, &logs)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("healthcheck did not succeed: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned %v after shutdown, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the server did not shut down")
	}
	if !bytes.Contains([]byte(logs.String()), []byte(`"path":"/healthz"`)) {
		t.Errorf("logs do not contain the health request: %s", logs.String())
	}
}

func TestHealthcheckFailsWhenServerIsDown(t *testing.T) {
	getenv := env(map[string]string{"PORT": freePort(t)})
	if err := run(t.Context(), []string{"healthcheck"}, getenv, &syncBuffer{}); err == nil {
		t.Fatal("healthcheck succeeded, want an error")
	}
}

func TestUnknownCommand(t *testing.T) {
	if err := run(t.Context(), []string{"unknown"}, env(nil), &syncBuffer{}); err == nil {
		t.Fatal("run succeeded, want an error")
	}
}

func TestServerFailsOnBadPort(t *testing.T) {
	getenv := env(map[string]string{"PORT": "not-a-port"})
	if err := run(t.Context(), nil, getenv, &syncBuffer{}); err == nil {
		t.Fatal("run succeeded, want an error")
	}
}
