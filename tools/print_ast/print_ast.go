package main

import (
	"fmt"

	i "github.com/cgrunewald/golox/interpreter"
)

func main() {
	expr := i.Binary{
		Left: &i.Literal{
			Value: 1,
		},
		Operator: i.NewToken(i.TK_MINUS, "-", nil, 1),
		Right: &i.Literal{
			Value: 2,
		},
	}

	printer := i.NewASTPrinter()
	fmt.Println(printer.Print(&expr))
}
