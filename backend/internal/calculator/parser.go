package calculator

// parser is a precedence-climbing parser. It reads the symbols, the
// precedence and the associativity of the operations from the registry.
type parser struct {
	tokens []token
	next   int
	depth  int
}

func parse(tokens []token) (node, error) {
	p := &parser{tokens: tokens}
	tree, err := p.expression(0)
	if err != nil {
		return nil, err
	}
	switch tok := p.peek(); tok.kind {
	case tokenEnd:
		return tree, nil
	case tokenCloseBracket:
		return nil, syntaxError(tok.position, "a closing bracket has no opening bracket")
	default:
		return nil, syntaxError(tok.position, "an operator is missing")
	}
}

func (p *parser) peek() token { return p.tokens[p.next] }

func (p *parser) advance() token {
	tok := p.tokens[p.next]
	if tok.kind != tokenEnd {
		p.next++
	}
	return tok
}

// expression reads an operand and then all operators with a precedence of
// minPrecedence or more.
func (p *parser) expression(minPrecedence int) (node, error) {
	left, err := p.operand()
	if err != nil {
		return nil, err
	}
	for {
		tok := p.peek()
		if tok.kind != tokenSymbol {
			return left, nil
		}
		if op := lookup(tok.text, Postfix); op != nil && op.Precedence >= minPrecedence {
			p.advance()
			left = &unaryNode{op: op, operand: left}
			continue
		}
		op := lookup(tok.text, Infix)
		if op == nil || op.Precedence < minPrecedence {
			return left, nil
		}
		p.advance()
		nextPrecedence := op.Precedence + 1
		if op.RightAssociative {
			nextPrecedence = op.Precedence
		}
		right, err := p.expression(nextPrecedence)
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: op, left: left, right: right}
	}
}

// operand reads a number, a bracket group, or a prefix operator with its operand.
func (p *parser) operand() (node, error) {
	tok := p.advance()
	switch tok.kind {
	case tokenNumber:
		return &numberNode{text: tok.text, value: tok.value}, nil
	case tokenOpenBracket:
		return p.group(tok)
	case tokenSymbol:
		if tok.text == symbolMinus {
			operand, err := p.expression(unaryMinusPrecedence)
			if err != nil {
				return nil, err
			}
			return &negateNode{operand: operand}, nil
		}
		if op := lookup(tok.text, Prefix); op != nil {
			operand, err := p.expression(op.Precedence)
			if err != nil {
				return nil, err
			}
			return &unaryNode{op: op, operand: operand}, nil
		}
	case tokenCloseBracket, tokenEnd:
	}
	return nil, syntaxError(tok.position, "a number is missing")
}

func (p *parser) group(open token) (node, error) {
	p.depth++
	if p.depth > MaxBracketDepth {
		return nil, syntaxError(open.position, "the brackets are nested too deeply")
	}
	inner, err := p.expression(0)
	if err != nil {
		return nil, err
	}
	switch tok := p.peek(); tok.kind {
	case tokenCloseBracket:
		p.advance()
	case tokenEnd:
		return nil, syntaxError(open.position, "a closing bracket is missing")
	default:
		return nil, syntaxError(tok.position, "an operator is missing")
	}
	p.depth--
	return &groupNode{inner: inner}, nil
}
