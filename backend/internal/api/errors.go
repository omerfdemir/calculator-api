package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"calculator-api/backend/internal/calculator"
)

// errorCode is an error code and its HTTP status.
type errorCode struct {
	name   string
	status int
}

// Error codes. They are part of the public contract.
var (
	codeInvalidRequest    = errorCode{"INVALID_REQUEST", http.StatusBadRequest}
	codeRequestTooLarge   = errorCode{"REQUEST_TOO_LARGE", http.StatusRequestEntityTooLarge}
	codeInvalidExpression = errorCode{"INVALID_EXPRESSION", http.StatusBadRequest}
	codeDivisionByZero    = errorCode{"DIVISION_BY_ZERO", http.StatusUnprocessableEntity}
	codeInvalidDomain     = errorCode{"INVALID_DOMAIN", http.StatusUnprocessableEntity}
	codeResultOutOfRange  = errorCode{"RESULT_OUT_OF_RANGE", http.StatusUnprocessableEntity}
	codeNotFound          = errorCode{"NOT_FOUND", http.StatusNotFound}
	codeMethodNotAllowed  = errorCode{"METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed}
	codeInternalError     = errorCode{"INTERNAL_ERROR", http.StatusInternalServerError}
)

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Position *int   `json:"position,omitempty"`
}

// domainErrors maps each domain error to its error code.
var domainErrors = []struct {
	err  error
	code errorCode
}{
	{calculator.ErrDivisionByZero, codeDivisionByZero},
	{calculator.ErrInvalidDomain, codeInvalidDomain},
	{calculator.ErrResultOutOfRange, codeResultOutOfRange},
}

func writeCalculatorError(w http.ResponseWriter, err error) {
	var syntaxErr *calculator.SyntaxError
	if errors.As(err, &syntaxErr) {
		writeError(w, codeInvalidExpression, syntaxErr.Reason, &syntaxErr.Position)
		return
	}
	for _, d := range domainErrors {
		if errors.Is(err, d.err) {
			writeError(w, d.code, d.err.Error(), nil)
			return
		}
	}
	writeInternalError(w)
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, codeInternalError, "an internal error occurred", nil)
}

func writeError(w http.ResponseWriter, code errorCode, message string, position *int) {
	writeJSON(w, code.status, errorResponse{Error: errorDetail{Code: code.name, Message: message, Position: position}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // The client is gone if the write fails.
}
