package interpreter

import (
	"testing"
)

func TestAST1(t *testing.T) {
	expr := Binary{
		Left: &Literal{
			Value: 1,
		},
		Operator: NewToken(TK_MINUS, "-", nil, 1),
		Right: &Literal{
			Value: 2,
		},
	}

	printer := NewASTPrinter()
	if str := printer.Print(&expr); str != "(- 1 2)" {
		t.Errorf("Expected (- 1 2), got %s", str)
	}

}

func TestAST2(t *testing.T) {
	expr := Binary{
		Left: &Literal{
			Value: 1,
		},
		Operator: NewToken(TK_MINUS, "-", nil, 1),
		Right: &Unary{
			Operator: NewToken(TK_MINUS, "-", nil, 1),
			Right: &Literal{
				Value: 2,
			},
		},
	}

	printer := NewASTPrinter()
	if str := printer.Print(&expr); str != "(- 1 (- 2))" {
		t.Errorf("Expected (- 1 (- 2)), got %s", str)
	}

}

func TestAST3(t *testing.T) {
	expr := Binary{
		Left: &Literal{
			Value: 1,
		},
		Operator: NewToken(TK_STAR, "*", nil, 1),
		Right: &Grouping{
			Expression: &Binary{
				Left: &Literal{
					Value: 3,
				},
				Operator: NewToken(TK_MINUS, "-", nil, 1),
				Right: &Literal{
					Value: 4,
				},
			},
		},
	}

	printer := NewASTPrinter()
	if printer.Print(&expr) != "(* 1 (group (- 3 4)))" {
		t.Errorf("Expected (* 1 (group (- 3 4))) , got %s", printer.Print(&expr))
	}

}
