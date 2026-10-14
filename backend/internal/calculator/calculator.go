// Package calculator parses and evaluates calculator expressions.
//
// The package has no HTTP code. The registry holds the operations, the parser
// reads their symbols and precedence from the registry, and the evaluator
// calls their functions.
package calculator

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// Limits of an expression.
const (
	MaxLength       = 200
	MaxBracketDepth = 32
)

// Result is the normalized expression and its value.
type Result struct {
	Expression string
	Value      float64
}

// Domain errors. The expression is valid, but it has no mathematical result.
var (
	ErrDivisionByZero   = errors.New("cannot divide by zero")
	ErrInvalidDomain    = errors.New("input is outside the domain of the operation")
	ErrResultOutOfRange = errors.New("result is out of range")
)

// SyntaxError is an expression that cannot be parsed. Position is the 0-based
// index of a Unicode code point in the expression.
type SyntaxError struct {
	Position int
	Reason   string
}

func (e *SyntaxError) Error() string { return e.Reason }

func syntaxError(position int, reason string) *SyntaxError {
	return &SyntaxError{Position: position, Reason: reason}
}

// Evaluate parses and calculates an expression. It returns a *SyntaxError for
// an expression that cannot be parsed, or one of the domain errors.
func Evaluate(expression string) (Result, error) {
	if utf8.RuneCountInString(expression) > MaxLength {
		return Result{}, syntaxError(MaxLength, "the expression is too long")
	}
	if strings.TrimSpace(expression) == "" {
		return Result{}, syntaxError(0, "the expression is empty")
	}

	tokens, err := lex(expression)
	if err != nil {
		return Result{}, err
	}
	tree, err := parse(tokens)
	if err != nil {
		return Result{}, err
	}
	value, err := tree.eval()
	if err != nil {
		return Result{}, err
	}
	if value == 0 {
		value = 0 // Turn -0 into 0.
	}
	return Result{Expression: tree.String(), Value: value}, nil
}
