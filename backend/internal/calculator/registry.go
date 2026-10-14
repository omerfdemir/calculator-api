package calculator

import (
	"math"
	"slices"
)

// Notation tells where the symbol of an operation goes.
type Notation string

// The notations of an operation.
const (
	Infix   Notation = "infix"
	Prefix  Notation = "prefix"
	Postfix Notation = "postfix"
)

// Operation is 1 entry of the registry. A higher precedence binds more tightly.
type Operation struct {
	Name             string
	Symbol           string
	Notation         Notation
	Arity            int
	Precedence       int
	RightAssociative bool
	apply            func(operands ...float64) (float64, error)
}

// unaryMinusPrecedence is the precedence of unary minus. It is a syntax rule,
// not an operation, so it is not in the registry.
const unaryMinusPrecedence = 3

// registry holds the operations in the fixed list order. To add an operation,
// add 1 entry.
var registry = []Operation{
	{Name: "add", Symbol: "+", Notation: Infix, Arity: 2, Precedence: 1,
		apply: func(x ...float64) (float64, error) { return x[0] + x[1], nil }},
	{Name: "subtract", Symbol: "−", Notation: Infix, Arity: 2, Precedence: 1,
		apply: func(x ...float64) (float64, error) { return x[0] - x[1], nil }},
	{Name: "multiply", Symbol: "×", Notation: Infix, Arity: 2, Precedence: 2,
		apply: func(x ...float64) (float64, error) { return x[0] * x[1], nil }},
	{Name: "divide", Symbol: "÷", Notation: Infix, Arity: 2, Precedence: 2, apply: divide},
	{Name: "power", Symbol: "^", Notation: Infix, Arity: 2, Precedence: 4, RightAssociative: true, apply: power},
	{Name: "sqrt", Symbol: "√", Notation: Prefix, Arity: 1, Precedence: 3, apply: sqrt},
	{Name: "percentage", Symbol: "%", Notation: Postfix, Arity: 1, Precedence: 5,
		apply: func(x ...float64) (float64, error) { return x[0] / 100, nil }},
}

// Operations returns a copy of the registry in its fixed order.
func Operations() []Operation {
	return slices.Clone(registry)
}

func divide(x ...float64) (float64, error) {
	if x[1] == 0 {
		return 0, ErrDivisionByZero
	}
	return x[0] / x[1], nil
}

func power(x ...float64) (float64, error) {
	if x[0] == 0 && x[1] < 0 {
		return 0, ErrDivisionByZero
	}
	result := math.Pow(x[0], x[1])
	if math.IsNaN(result) {
		return 0, ErrInvalidDomain
	}
	return result, nil
}

func sqrt(x ...float64) (float64, error) {
	if x[0] < 0 {
		return 0, ErrInvalidDomain
	}
	return math.Sqrt(x[0]), nil
}

// lookup returns the operation with this symbol and notation, or nil.
func lookup(symbol string, notation Notation) *Operation {
	for i := range registry {
		if registry[i].Symbol == symbol && registry[i].Notation == notation {
			return &registry[i]
		}
	}
	return nil
}

// isSymbol reports whether s is the symbol of an operation.
func isSymbol(s string) bool {
	return slices.ContainsFunc(registry, func(op Operation) bool { return op.Symbol == s })
}
