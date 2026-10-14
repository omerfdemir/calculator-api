package calculator_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"calculator-api/backend/internal/calculator"
)

func TestOperationsListIsFixed(t *testing.T) {
	want := []struct {
		name     string
		symbol   string
		notation calculator.Notation
		arity    int
	}{
		{"add", "+", calculator.Infix, 2},
		{"subtract", "−", calculator.Infix, 2},
		{"multiply", "×", calculator.Infix, 2},
		{"divide", "÷", calculator.Infix, 2},
		{"power", "^", calculator.Infix, 2},
		{"sqrt", "√", calculator.Prefix, 1},
		{"percentage", "%", calculator.Postfix, 1},
	}

	got := calculator.Operations()
	if len(got) != len(want) {
		t.Fatalf("got %d operations, want %d", len(got), len(want))
	}
	for i, w := range want {
		g := got[i]
		if g.Name != w.name || g.Symbol != w.symbol || g.Notation != w.notation || g.Arity != w.arity {
			t.Errorf("operation %d = {%s %s %s %d}, want {%s %s %s %d}",
				i, g.Name, g.Symbol, g.Notation, g.Arity, w.name, w.symbol, w.notation, w.arity)
		}
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     float64
		wantExpr string // empty means: do not check the normalized expression
	}{
		// One operation each.
		{"add", "2+3", 5, "2 + 3"},
		{"subtract", "10−4", 6, "10 − 4"},
		{"multiply", "6×7", 42, "6 × 7"},
		{"divide", "1÷4", 0.25, "1 ÷ 4"},
		{"power", "2^10", 1024, "2 ^ 10"},
		{"sqrt", "√9", 3, "√9"},
		{"percentage", "50%", 0.5, "50%"},

		// Numbers.
		{"negative decimal", "−1.5×2", -3, "−1.5 × 2"},
		{"leading dot", ".5+.5", 1, ".5 + .5"},
		{"exponent notation", "1e6+1", 1000001, "1e6 + 1"},
		{"negative exponent ascii", "2.5e-3", 0.0025, "2.5e-3"},
		{"negative exponent unicode", "2.5E−3", 0.0025, "2.5E−3"},
		{"ascii symbols", "2*3/4-1", 0.5, "2 × 3 ÷ 4 − 1"},
		{"spaces", "  2 +   3 ", 5, "2 + 3"},

		// Precedence, associativity and brackets.
		{"multiply before add", "2+3×4", 14, "2 + 3 × 4"},
		{"brackets", "(2+3)×4", 20, "(2 + 3) × 4"},
		{"power is right-associative", "2^3^2", 512, "2 ^ 3 ^ 2"},
		{"power before unary minus", "−2^2", -4, "−2 ^ 2"},
		{"power before sqrt", "√4^2", 4, "√4 ^ 2"},
		{"sqrt before multiply", "√4×2", 4, "√4 × 2"},
		{"subtract is left-associative", "10−4−3", 3, "10 − 4 − 3"},
		{"divide is left-associative", "8÷4÷2", 1, "8 ÷ 4 ÷ 2"},
		{"unary minus after operator", "2×−3", -6, "2 × −3"},
		{"negative exponent", "2^−1", 0.5, "2 ^ −1"},
		{"unary minus before brackets", "−(2+3)", -5, "−(2 + 3)"},
		{"double unary minus", "−−5", 5, "−−5"},
		{"nested brackets", "((1+2)×(3+4))", 21, "((1 + 2) × (3 + 4))"},
		{"sqrt of brackets", "√(2+7)", 3, "√(2 + 7)"},
		{"bracketed sqrt to a power", "(√4)^3", 8, "(√4) ^ 3"},

		// Phone percentage rule.
		{"percent of a number", "15%×80", 12, "15% × 80"},
		{"add a percent", "80+15%", 92, "80 + 15%"},
		{"subtract a percent", "80−15%", 68, "80 − 15%"},
		{"add a percent twice", "100+10%+10%", 121, "100 + 10% + 10%"},
		{"percent term is not direct", "200+10%×2", 200.2, "200 + 10% × 2"},
		{"bracketed percent is not direct", "80+(15%)", 80.15, "80 + (15%)"},
		{"percent before power", "50%^2", 0.25, "50% ^ 2"},
		{"percent of a large number", "1e308+50%", 1.5e308, ""},

		// Numeric edge cases.
		{"raw float64 result", "0.1+0.2", 0.30000000000000004, ""},
		{"very large exponent", "1e308", 1e308, ""},
		{"very small exponent", "1e−308", 1e-308, ""},
		{"largest power of 2", "2^1023", 8.98846567431158e307, ""},
		{"smallest subnormal", "2^−1074", 5e-324, ""},
		{"underflow to zero", "1e−300×1e−300", 0, ""},
		{"very large literal", "1.7976931348623157e308", math.MaxFloat64, ""},
		{"maximum bracket depth", strings.Repeat("(", 32) + "1" + strings.Repeat(")", 32), 1, ""},
		{"maximum length", strings.Repeat("1+", 99) + "1 ", 100, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.Evaluate(tt.input)
			if err != nil {
				t.Fatalf("Evaluate(%q) error = %v", tt.input, err)
			}
			if got.Value != tt.want {
				t.Errorf("Evaluate(%q) = %v, want %v", tt.input, got.Value, tt.want)
			}
			if tt.wantExpr != "" && got.Expression != tt.wantExpr {
				t.Errorf("Evaluate(%q) expression = %q, want %q", tt.input, got.Expression, tt.wantExpr)
			}
		})
	}
}

func TestEvaluateNormalizesNegativeZero(t *testing.T) {
	for _, input := range []string{"−0", "0×−1", "−0÷5", "√−0"} {
		t.Run(input, func(t *testing.T) {
			got, err := calculator.Evaluate(input)
			if err != nil {
				t.Fatalf("Evaluate(%q) error = %v", input, err)
			}
			if got.Value != 0 || math.Signbit(got.Value) {
				t.Errorf("Evaluate(%q) = %v (sign bit %t), want +0", input, got.Value, math.Signbit(got.Value))
			}
		})
	}
}

func TestEvaluateDomainErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"division by zero", "1÷0", calculator.ErrDivisionByZero},
		{"division by negative zero", "1÷−0", calculator.ErrDivisionByZero},
		{"zero to a negative power", "0^−1", calculator.ErrDivisionByZero},
		{"negative base, fractional exponent", "(−8)^0.5", calculator.ErrInvalidDomain},
		{"sqrt of a negative number", "√−1", calculator.ErrInvalidDomain},
		{"multiply overflow", "1e308×10", calculator.ErrResultOutOfRange},
		{"negative overflow", "−1e308×10", calculator.ErrResultOutOfRange},
		{"add overflow", "1e308+1e308", calculator.ErrResultOutOfRange},
		{"power overflow", "2^1024", calculator.ErrResultOutOfRange},
		{"overflow in a step", "1e308×10×0", calculator.ErrResultOutOfRange},
		{"first error wins", "1÷0+√−1", calculator.ErrDivisionByZero},
		{"error inside sqrt", "√(1÷0)", calculator.ErrDivisionByZero},
		{"error in right operand", "2+√−4", calculator.ErrInvalidDomain},
		{"overflow in percentage rule", "1e308+1e308%", calculator.ErrResultOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calculator.Evaluate(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Evaluate(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestEvaluateSyntaxErrors(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantPosition int
	}{
		{"empty", "", 0},
		{"only spaces", "   ", 0},
		{"missing right operand", "2+", 2},
		{"unary plus", "+2", 0},
		{"two infix operators", "2++3", 2},
		{"missing closing bracket", "(2+3", 0},
		{"missing inner closing bracket", "2×((1+2)", 2},
		{"bracket without a pair", "2+3)", 3},
		{"empty brackets", "()", 1},
		{"implicit multiplication with brackets", "2(3)", 1},
		{"implicit multiplication with sqrt", "2√9", 1},
		{"two numbers", "2 3", 2},
		{"unknown character", "2a", 1},
		{"position counts characters, not bytes", "√9+a", 3},
		{"sqrt without operand", "√", 1},
		{"unary minus without operand", "2×−", 3},
		{"two numbers in brackets", "(2 3)", 3},
		{"percent without operand", "%", 0},
		{"exponent without digits", "1e", 0},
		{"exponent sign without digits", "3+1e+", 2},
		{"lone dot", ".", 0},
		{"two decimal points", "1.2.3", 3},
		{"literal not finite", "2×1e999", 2},
		{"too long", strings.Repeat("1", 201), 200},
		{"too deep", strings.Repeat("(", 33) + "1" + strings.Repeat(")", 33), 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calculator.Evaluate(tt.input)
			var syntaxErr *calculator.SyntaxError
			if !errors.As(err, &syntaxErr) {
				t.Fatalf("Evaluate(%q) error = %v, want a *SyntaxError", tt.input, err)
			}
			if syntaxErr.Position != tt.wantPosition {
				t.Errorf("Evaluate(%q) position = %d, want %d (%v)",
					tt.input, syntaxErr.Position, tt.wantPosition, syntaxErr)
			}
			if syntaxErr.Error() == "" {
				t.Errorf("Evaluate(%q) error message is empty", tt.input)
			}
		})
	}
}
