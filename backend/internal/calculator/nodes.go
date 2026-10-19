package calculator

import "math"

// node is 1 part of the syntax tree. String returns the normalized form.
type node interface {
	eval() (float64, error)
	String() string
}

type numberNode struct {
	text  string
	value float64
}

func (n *numberNode) eval() (float64, error) { return n.value, nil }
func (n *numberNode) String() string         { return n.text }

type groupNode struct{ inner node }

func (n *groupNode) eval() (float64, error) { return n.inner.eval() }
func (n *groupNode) String() string         { return "(" + n.inner.String() + ")" }

type negateNode struct{ operand node }

func (n *negateNode) eval() (float64, error) {
	value, err := n.operand.eval()
	return -value, err
}

func (n *negateNode) String() string { return symbolMinus + n.operand.String() }

// unaryNode is a prefix or a postfix operation.
type unaryNode struct {
	op      *Operation
	operand node
}

func (n *unaryNode) eval() (float64, error) {
	value, err := n.operand.eval()
	if err != nil {
		return 0, err
	}
	return finite(n.op.apply(value))
}

func (n *unaryNode) String() string {
	if n.op.Notation == Postfix {
		return n.operand.String() + n.op.Symbol
	}
	return n.op.Symbol + n.operand.String()
}

type binaryNode struct {
	op          *Operation
	left, right node
}

func (n *binaryNode) eval() (float64, error) {
	left, err := n.left.eval()
	if err != nil {
		return 0, err
	}
	right, err := n.right.eval()
	if err != nil {
		return 0, err
	}
	if n.isPercentOfLeft() {
		// Phone percentage rule: in a + b% and a − b%, the percentage is of a.
		if right, err = finite(left*right, nil); err != nil {
			return 0, err
		}
	}
	return finite(n.op.apply(left, right))
}

// isPercentOfLeft reports whether the right operand of + or − is directly a
// % term, with or without unary minus signs before it (80 + −15%).
func (n *binaryNode) isPercentOfLeft() bool {
	if n.op.Name != nameAdd && n.op.Name != nameSubtract {
		return false
	}
	right := n.right
	for {
		negate, ok := right.(*negateNode)
		if !ok {
			break
		}
		right = negate.operand
	}
	percent, ok := right.(*unaryNode)
	return ok && percent.op.Name == namePercentage
}

func (n *binaryNode) String() string {
	return n.left.String() + " " + n.op.Symbol + " " + n.right.String()
}

// finite turns an infinite value into ErrResultOutOfRange.
func finite(value float64, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	if math.IsInf(value, 0) {
		return 0, ErrResultOutOfRange
	}
	return value, nil
}
