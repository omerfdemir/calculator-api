package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"calculator-api/backend/internal/calculator"
)

type handlers struct {
	evaluate func(string) (calculator.Result, error)
}

type calculateResponse struct {
	Expression string  `json:"expression"`
	Result     float64 `json:"result"`
}

type operationResponse struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Notation string `json:"notation"`
	Arity    int    `json:"arity"`
}

func (h *handlers) calculate(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, codeInvalidRequest, "the Content-Type must be application/json", nil)
		return
	}

	expression, err := decodeExpression(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, codeRequestTooLarge, "the request body is larger than 1 MB", nil)
			return
		}
		writeError(w, codeInvalidRequest, "the request body is not a valid calculate request", nil)
		return
	}
	if expression == nil {
		writeError(w, codeInvalidRequest, "the expression is required and must be a string", nil)
		return
	}

	result, err := h.evaluate(*expression)
	if err != nil {
		writeCalculatorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, calculateResponse{Expression: result.Expression, Result: result.Value})
}

var errInvalidBody = errors.New("the body is not a valid calculate request")

// decodeExpression reads exactly 1 JSON object. The only allowed key is
// "expression", exactly once and in this case. encoding/json matches keys
// without case and keeps the last duplicate, so the keys are read as tokens.
// A missing or null expression returns nil.
func decodeExpression(body io.Reader) (*string, error) {
	dec := json.NewDecoder(body)
	if tok, err := dec.Token(); err != nil {
		return nil, err
	} else if tok != json.Delim('{') {
		return nil, errInvalidBody
	}

	var expression *string
	seen := false
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if tok != "expression" || seen {
			return nil, errInvalidBody
		}
		seen = true
		if err := dec.Decode(&expression); err != nil {
			return nil, err
		}
	}
	if _, err := dec.Token(); err != nil { // The closing brace.
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) { // Nothing after the object.
		if err == nil {
			return nil, errInvalidBody
		}
		return nil, err
	}
	return expression, nil
}

func (h *handlers) operations(w http.ResponseWriter, _ *http.Request) {
	ops := calculator.Operations()
	resp := make([]operationResponse, 0, len(ops))
	for _, op := range ops {
		resp = append(resp, operationResponse{
			Name: op.Name, Symbol: op.Symbol, Notation: string(op.Notation), Arity: op.Arity,
		})
	}
	writeJSON(w, http.StatusOK, map[string][]operationResponse{"operations": resp})
}

func (h *handlers) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, codeNotFound, "the path does not exist", nil)
}

func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, codeMethodNotAllowed, "the method is not allowed", nil)
	}
}
