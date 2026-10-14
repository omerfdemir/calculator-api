package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"calculator-api/backend/internal/calculator"
)

// Error codes. They are part of the public contract.
const (
	codeInvalidRequest    = "INVALID_REQUEST"
	codeRequestTooLarge   = "REQUEST_TOO_LARGE"
	codeInvalidExpression = "INVALID_EXPRESSION"
	codeDivisionByZero    = "DIVISION_BY_ZERO"
	codeInvalidDomain     = "INVALID_DOMAIN"
	codeResultOutOfRange  = "RESULT_OUT_OF_RANGE"
	codeMethodNotAllowed  = "METHOD_NOT_ALLOWED"
	codeInternalError     = "INTERNAL_ERROR"
)

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Position *int   `json:"position,omitempty"`
}

// domainErrors maps each domain error to its error code. All get status 422.
var domainErrors = []struct {
	err  error
	code string
}{
	{calculator.ErrDivisionByZero, codeDivisionByZero},
	{calculator.ErrInvalidDomain, codeInvalidDomain},
	{calculator.ErrResultOutOfRange, codeResultOutOfRange},
}

func writeCalculatorError(w http.ResponseWriter, err error) {
	var syntaxErr *calculator.SyntaxError
	if errors.As(err, &syntaxErr) {
		writeError(w, http.StatusBadRequest, codeInvalidExpression, syntaxErr.Reason, &syntaxErr.Position)
		return
	}
	for _, d := range domainErrors {
		if errors.Is(err, d.err) {
			writeError(w, http.StatusUnprocessableEntity, d.code, d.err.Error(), nil)
			return
		}
	}
	writeError(w, http.StatusInternalServerError, codeInternalError, "an internal error occurred", nil)
}

func writeError(w http.ResponseWriter, status int, code, message string, position *int) {
	writeJSON(w, status, errorResponse{Error: errorDetail{Code: code, Message: message, Position: position}})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // The client is gone if the write fails.
}
