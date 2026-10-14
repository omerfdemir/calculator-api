package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"calculator-api/backend/internal/api"
	"calculator-api/backend/internal/calculator"
)

const devOrigin = "http://localhost:5173"

func newHandler(t *testing.T) http.Handler {
	t.Helper()
	return api.NewHandler(api.Config{
		AllowedOrigins: []string{devOrigin},
		Logger:         slog.New(slog.DiscardHandler),
	})
}

type request struct {
	method      string
	path        string
	contentType string
	body        string
	headers     map[string]string
}

func send(t *testing.T, h http.Handler, r request) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), r.method, r.path, strings.NewReader(r.body))
	if r.contentType != "" {
		req.Header.Set("Content-Type", r.contentType)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func calculate(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	return send(t, h, request{method: http.MethodPost, path: "/api/v1/calculate", contentType: "application/json", body: body})
}

// assertOK checks the status 200, the JSON content type and that the body is equal to want as JSON.
func assertOK(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var got, exp any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not JSON: %v (%s)", err, rec.Body)
	}
	if err := json.Unmarshal([]byte(want), &exp); err != nil {
		t.Fatalf("want is not JSON: %v", err)
	}
	if !reflect.DeepEqual(got, exp) {
		t.Errorf("body = %s, want %s", rec.Body, want)
	}
}

type errorBody struct {
	Error struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Position *int   `json:"position"`
	} `json:"error"`
}

// assertError checks the status, the error code, a non-empty message and the position.
func assertError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantCode string, wantPosition *int) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Errorf("status = %d, want %d (body %s)", rec.Code, wantStatus, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body errorBody
	dec := json.NewDecoder(rec.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		t.Fatalf("error body has the wrong shape: %v", err)
	}
	if body.Error.Code != wantCode {
		t.Errorf("code = %q, want %q", body.Error.Code, wantCode)
	}
	if body.Error.Message == "" {
		t.Error("message is empty")
	}
	switch {
	case wantPosition == nil && body.Error.Position != nil:
		t.Errorf("position = %d, want no position", *body.Error.Position)
	case wantPosition != nil && body.Error.Position == nil:
		t.Errorf("position is missing, want %d", *wantPosition)
	case wantPosition != nil && *body.Error.Position != *wantPosition:
		t.Errorf("position = %d, want %d", *body.Error.Position, *wantPosition)
	}
}

func position(p int) *int { return &p }

func TestCalculate(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"add", `{"expression":"2+3"}`, `{"expression":"2 + 3","result":5}`},
		{"subtract", `{"expression":"10−4"}`, `{"expression":"10 − 4","result":6}`},
		{"multiply", `{"expression":"6×7"}`, `{"expression":"6 × 7","result":42}`},
		{"divide", `{"expression":"1÷4"}`, `{"expression":"1 ÷ 4","result":0.25}`},
		{"power", `{"expression":"2^10"}`, `{"expression":"2 ^ 10","result":1024}`},
		{"sqrt", `{"expression":"√9"}`, `{"expression":"√9","result":3}`},
		{"percentage", `{"expression":"15%×80"}`, `{"expression":"15% × 80","result":12}`},
		{"phone percentage", `{"expression":"80+15%"}`, `{"expression":"80 + 15%","result":92}`},
		{"precedence", `{"expression":"2+3*4"}`, `{"expression":"2 + 3 × 4","result":14}`},
		{"brackets", `{"expression":"(2+3)*4"}`, `{"expression":"(2 + 3) × 4","result":20}`},
		{"negative and decimal", `{"expression":"-1.5*2"}`, `{"expression":"−1.5 × 2","result":-3}`},
		{"exponent notation", `{"expression":"1e6+1"}`, `{"expression":"1e6 + 1","result":1000001}`},
	}

	h := newHandler(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOK(t, calculate(t, h, tt.body), tt.want)
		})
	}
}

func TestCalculateReturnsRawFloat64(t *testing.T) {
	rec := calculate(t, newHandler(t), `{"expression":"0.1+0.2"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"result":0.30000000000000004`) {
		t.Errorf("body = %s, want the raw float64 value 0.30000000000000004", rec.Body)
	}
}

func TestCalculateAcceptsJSONWithCharset(t *testing.T) {
	rec := send(t, newHandler(t), request{
		method: http.MethodPost, path: "/api/v1/calculate",
		contentType: "application/json; charset=utf-8", body: `{"expression":"1+1"}`,
	})
	assertOK(t, rec, `{"expression":"1 + 1","result":2}`)
}

func TestCalculateErrors(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		body         string
		wantStatus   int
		wantCode     string
		wantPosition *int
	}{
		// INVALID_REQUEST
		{"malformed JSON", "application/json", `{"expression":`, 400, "INVALID_REQUEST", nil},
		{"empty body", "application/json", ``, 400, "INVALID_REQUEST", nil},
		{"unknown field", "application/json", `{"expression":"1+1","extra":1}`, 400, "INVALID_REQUEST", nil},
		{"key in other case", "application/json", `{"Expression":"1+1"}`, 400, "INVALID_REQUEST", nil},
		{"key in upper case", "application/json", `{"EXPRESSION":"1+1"}`, 400, "INVALID_REQUEST", nil},
		{"duplicate key", "application/json", `{"expression":"1/0","expression":"1+1"}`, 400, "INVALID_REQUEST", nil},
		{"duplicate key in other case", "application/json", `{"expression":"1+1","EXPRESSION":"2"}`, 400, "INVALID_REQUEST", nil},
		{"key is not a string", "application/json", `{1:"1+1"}`, 400, "INVALID_REQUEST", nil},
		{"missing closing brace", "application/json", `{"expression":"1+1"`, 400, "INVALID_REQUEST", nil},
		{"text after the object", "application/json", `{"expression":"1+1"} x`, 400, "INVALID_REQUEST", nil},
		{"missing expression", "application/json", `{}`, 400, "INVALID_REQUEST", nil},
		{"null expression", "application/json", `{"expression":null}`, 400, "INVALID_REQUEST", nil},
		{"expression is a number", "application/json", `{"expression":5}`, 400, "INVALID_REQUEST", nil},
		{"JSON array", "application/json", `[]`, 400, "INVALID_REQUEST", nil},
		{"two JSON values", "application/json", `{"expression":"1"}{"expression":"2"}`, 400, "INVALID_REQUEST", nil},
		{"second JSON value is empty", "application/json", `{"expression":"1"} {}`, 400, "INVALID_REQUEST", nil},
		{"content type is not JSON", "text/plain", `{"expression":"1+1"}`, 400, "INVALID_REQUEST", nil},
		{"content type is missing", "", `{"expression":"1+1"}`, 400, "INVALID_REQUEST", nil},
		{"content type is malformed", "application/json; =", `{"expression":"1+1"}`, 400, "INVALID_REQUEST", nil},

		// INVALID_EXPRESSION
		{"empty expression", "application/json", `{"expression":""}`, 400, "INVALID_EXPRESSION", position(0)},
		{"missing operand", "application/json", `{"expression":"2+"}`, 400, "INVALID_EXPRESSION", position(2)},
		{"missing closing bracket", "application/json", `{"expression":"2×(1+2"}`, 400, "INVALID_EXPRESSION", position(2)},
		{"unknown character", "application/json", `{"expression":"√9+a"}`, 400, "INVALID_EXPRESSION", position(3)},
		{"too long", "application/json", `{"expression":"` + strings.Repeat("1", 201) + `"}`, 400, "INVALID_EXPRESSION", position(200)},

		// Domain errors
		{"division by zero", "application/json", `{"expression":"1/0"}`, 422, "DIVISION_BY_ZERO", nil},
		{"zero to a negative power", "application/json", `{"expression":"0^-1"}`, 422, "DIVISION_BY_ZERO", nil},
		{"sqrt of a negative number", "application/json", `{"expression":"√-1"}`, 422, "INVALID_DOMAIN", nil},
		{"negative base, fractional exponent", "application/json", `{"expression":"(-8)^0.5"}`, 422, "INVALID_DOMAIN", nil},
		{"result out of range", "application/json", `{"expression":"1e308*10"}`, 422, "RESULT_OUT_OF_RANGE", nil},
	}

	h := newHandler(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := send(t, h, request{method: http.MethodPost, path: "/api/v1/calculate", contentType: tt.contentType, body: tt.body})
			assertError(t, rec, tt.wantStatus, tt.wantCode, tt.wantPosition)
		})
	}
}

func TestRequestTooLarge(t *testing.T) {
	body := `{"expression":"` + strings.Repeat("1", 1<<20) + `"}`
	assertError(t, calculate(t, newHandler(t), body), http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", nil)
}

func TestMethodNotAllowed(t *testing.T) {
	tests := []struct {
		method    string
		path      string
		wantAllow string
	}{
		{http.MethodGet, "/api/v1/calculate", "POST"},
		{http.MethodPut, "/api/v1/calculate", "POST"},
		{http.MethodPost, "/api/v1/operations", "GET, HEAD"},
		{http.MethodDelete, "/healthz", "GET, HEAD"},
	}

	h := newHandler(t)
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := send(t, h, request{method: tt.method, path: tt.path})
			assertError(t, rec, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", nil)
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/unknown"},
		{http.MethodPost, "/api/v2/calculate"},
		{http.MethodGet, "/"},
		{http.MethodGet, "/healthz/extra"},
	}

	h := newHandler(t)
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := send(t, h, request{method: tt.method, path: tt.path})
			assertError(t, rec, http.StatusNotFound, "NOT_FOUND", nil)
		})
	}
}

func TestOperations(t *testing.T) {
	rec := send(t, newHandler(t), request{method: http.MethodGet, path: "/api/v1/operations"})
	assertOK(t, rec, `{"operations":[
		{"name":"add","symbol":"+","notation":"infix","arity":2},
		{"name":"subtract","symbol":"−","notation":"infix","arity":2},
		{"name":"multiply","symbol":"×","notation":"infix","arity":2},
		{"name":"divide","symbol":"÷","notation":"infix","arity":2},
		{"name":"power","symbol":"^","notation":"infix","arity":2},
		{"name":"sqrt","symbol":"√","notation":"prefix","arity":1},
		{"name":"percentage","symbol":"%","notation":"postfix","arity":1}
	]}`)
}

func TestHealthz(t *testing.T) {
	rec := send(t, newHandler(t), request{method: http.MethodGet, path: "/healthz"})
	assertOK(t, rec, `{"status":"ok"}`)
}

func TestCORS(t *testing.T) {
	h := newHandler(t)

	t.Run("allowed origin", func(t *testing.T) {
		rec := send(t, h, request{
			method: http.MethodPost, path: "/api/v1/calculate", contentType: "application/json",
			body: `{"expression":"1+1"}`, headers: map[string]string{"Origin": devOrigin},
		})
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != devOrigin {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, devOrigin)
		}
		assertVaryOrigin(t, rec)
	})

	t.Run("other origin", func(t *testing.T) {
		rec := send(t, h, request{
			method: http.MethodGet, path: "/api/v1/operations",
			headers: map[string]string{"Origin": "https://evil.example"},
		})
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, want no header", got)
		}
		assertVaryOrigin(t, rec)
	})

	t.Run("no origin", func(t *testing.T) {
		rec := send(t, h, request{method: http.MethodGet, path: "/healthz"})
		assertVaryOrigin(t, rec)
	})

	t.Run("preflight", func(t *testing.T) {
		rec := send(t, h, request{
			method: http.MethodOptions, path: "/api/v1/calculate",
			headers: map[string]string{"Origin": devOrigin, "Access-Control-Request-Method": "POST"},
		})
		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", rec.Code)
		}
		want := map[string]string{
			"Access-Control-Allow-Origin":  devOrigin,
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		}
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s = %q, want %q", k, got, v)
			}
		}
	})

	t.Run("preflight from other origin", func(t *testing.T) {
		rec := send(t, h, request{
			method: http.MethodOptions, path: "/api/v1/calculate",
			headers: map[string]string{"Origin": "https://evil.example", "Access-Control-Request-Method": "POST"},
		})
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, want no header", got)
		}
	})
}

// assertVaryOrigin checks that the response has exactly 1 Vary: Origin. A
// shared cache must not give a response for 1 origin to another origin.
func assertVaryOrigin(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Values("Vary"); len(got) != 1 || got[0] != "Origin" {
		t.Errorf("Vary = %q, want [Origin]", got)
	}
}

func TestPanicRecovery(t *testing.T) {
	h := api.NewHandler(api.Config{
		Logger: slog.New(slog.DiscardHandler),
		Evaluate: func(string) (calculator.Result, error) {
			panic("secret internal detail")
		},
	})
	rec := calculate(t, h, `{"expression":"1+1"}`)
	assertError(t, rec, http.StatusInternalServerError, "INTERNAL_ERROR", nil)
	if strings.Contains(rec.Body.String(), "secret") {
		t.Errorf("body shows internal details: %s", rec.Body)
	}
}

func TestUnexpectedErrorIsInternal(t *testing.T) {
	h := api.NewHandler(api.Config{
		Logger: slog.New(slog.DiscardHandler),
		Evaluate: func(string) (calculator.Result, error) {
			return calculator.Result{}, errors.New("secret internal detail")
		},
	})
	rec := calculate(t, h, `{"expression":"1+1"}`)
	assertError(t, rec, http.StatusInternalServerError, "INTERNAL_ERROR", nil)
	if strings.Contains(rec.Body.String(), "secret") {
		t.Errorf("body shows internal details: %s", rec.Body)
	}
}

func TestRequestLogging(t *testing.T) {
	var logs bytes.Buffer
	h := api.NewHandler(api.Config{Logger: slog.New(slog.NewTextHandler(&logs, nil))})
	send(t, h, request{method: http.MethodGet, path: "/healthz"})

	line := logs.String()
	for _, want := range []string{"method=GET", "path=/healthz", "status=200", "duration="} {
		if !strings.Contains(line, want) {
			t.Errorf("log %q does not contain %q", line, want)
		}
	}
}
