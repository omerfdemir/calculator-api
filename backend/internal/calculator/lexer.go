package calculator

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

type tokenKind int

const (
	tokenNumber tokenKind = iota
	tokenSymbol
	tokenOpenBracket
	tokenCloseBracket
	tokenEnd
)

type token struct {
	kind     tokenKind
	text     string // The number as written, or the Unicode symbol.
	value    float64
	position int
}

// asciiSymbols maps the ASCII symbols that the API also accepts.
var asciiSymbols = map[rune]string{'*': "×", '/': "÷", '-': symbolMinus}

func lex(expression string) ([]token, error) {
	runes := []rune(expression)
	var tokens []token
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case r == ' ' || r == '\t':
			i++
		case isDigit(r) || r == '.':
			tok, next, err := lexNumber(runes, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			i = next
		case r == '(':
			tokens = append(tokens, token{kind: tokenOpenBracket, text: "(", position: i})
			i++
		case r == ')':
			tokens = append(tokens, token{kind: tokenCloseBracket, text: ")", position: i})
			i++
		default:
			symbol, ok := asciiSymbols[r]
			if !ok {
				symbol = string(r)
			}
			if !isSymbol(symbol) {
				return nil, syntaxError(i, "unknown character "+strconv.Quote(string(r)))
			}
			tokens = append(tokens, token{kind: tokenSymbol, text: symbol, position: i})
			i++
		}
	}
	return append(tokens, token{kind: tokenEnd, position: len(runes)}), nil
}

// lexNumber reads digits, an optional decimal part and an optional exponent.
func lexNumber(runes []rune, start int) (token, int, error) {
	i := start
	digits := 0
	for i < len(runes) && isDigit(runes[i]) {
		i++
		digits++
	}
	if i < len(runes) && runes[i] == '.' {
		i++
		for i < len(runes) && isDigit(runes[i]) {
			i++
			digits++
		}
	}
	if digits == 0 {
		return token{}, 0, syntaxError(start, "a number has no digits")
	}
	if i < len(runes) && (runes[i] == 'e' || runes[i] == 'E') {
		i++
		if i < len(runes) && (runes[i] == '+' || runes[i] == '-' || runes[i] == '−') {
			i++
		}
		exponentDigits := 0
		for i < len(runes) && isDigit(runes[i]) {
			i++
			exponentDigits++
		}
		if exponentDigits == 0 {
			return token{}, 0, syntaxError(start, "an exponent has no digits")
		}
	}

	text := string(runes[start:i])
	value, err := strconv.ParseFloat(strings.ReplaceAll(text, symbolMinus, "-"), 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return token{}, 0, syntaxError(start, "a number is not valid")
	}
	if math.IsInf(value, 0) {
		return token{}, 0, syntaxError(start, "a number is too large")
	}
	return token{kind: tokenNumber, text: text, value: value, position: start}, i, nil
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }
