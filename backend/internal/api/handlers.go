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

type calculateRequest struct {
	Expression *string `json:"expression"`
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
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "the Content-Type must be application/json", nil)
		return
	}

	var req calculateRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, codeRequestTooLarge, "the request body is larger than 1 MB", nil)
			return
		}
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "the request body is not a valid calculate request", nil)
		return
	}
	if req.Expression == nil {
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "the expression is required and must be a string", nil)
		return
	}

	result, err := h.evaluate(*req.Expression)
	if err != nil {
		writeCalculatorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, calculateResponse{Expression: result.Expression, Result: result.Value})
}

// decodeJSON decodes exactly 1 JSON value with no unknown fields.
func decodeJSON(body io.Reader, v any) error {
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("the body has more than 1 JSON value")
		}
		return err
	}
	return nil
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

func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "the method is not allowed", nil)
	}
}
