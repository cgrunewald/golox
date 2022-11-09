package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

var exprGrammar []string = []string{
	"Binary : Left Expr, Operator Token, Right Expr",
	"Grouping : Expression Expr",
	"Literal : Value interface{}",
	"Unary : Operator Token, Right Expr",
	"TernaryCondition : Condition Expr, TrueBranch Expr, FalseBranch Expr",
}

type writeHelper struct {
	writer *bufio.Writer
	err    error
}

func (w *writeHelper) write(s string) {
	if w.err != nil {
		return
	}

	_, w.err = w.writer.WriteString(s)
}

func (w *writeHelper) writef(format string, args ...interface{}) {
	if w.err != nil {
		return
	}

	_, w.err = w.writer.WriteString(fmt.Sprintf(format, args...))
}

func (w *writeHelper) writeNewLine() {
	if w.err != nil {
		return
	}

	_, w.err = w.writer.WriteString("\n")
}

func (w *writeHelper) writeLinef(format string, args ...interface{}) {
	if w.err != nil {
		return
	}

	_, w.err = w.writer.WriteString(fmt.Sprintf(format, args...) + "\n")
}

func generate(writer *bufio.Writer, mainClass string, grammar []string) {
	w := &writeHelper{writer: writer}

	w.writeLinef("package interpreter")
	w.writeNewLine()
	w.writeLinef("type %s interface {", mainClass)
	w.writeLinef("  Accept(visitor %sVisitor) interface{}", mainClass)
	w.writeLinef("}")
	w.writeNewLine()

	visitorMethods := make([]string, 0, len(grammar))
	structDefiners := make([]func(), 0, len(grammar))

	for _, item := range grammar {
		prodcution := strings.Split(item, ":")
		key := strings.Trim(prodcution[0], " ")
		definition := strings.Trim(prodcution[1], " ")
		definitionElements := strings.Split(definition, ",")

		visitorMethods = append(visitorMethods, fmt.Sprintf("Visit%s(expr *%s) interface{}", key, key))

		definerFunc := func() {
			w.writeLinef("type %s struct {", key)
			w.writeLinef("  Expr")
			for _, def := range definitionElements {
				def = strings.Trim(def, " ")
				w.writeLinef("  %s", def)
			}
			w.writeLinef("}")
			w.writeNewLine()

			w.writeLinef("func (e *%s) Accept(visitor %sVisitor) interface{} {", key, mainClass)
			w.writeLinef("  return visitor.Visit%s(e)", key)
			w.writeLinef("}")
			w.writeNewLine()
		}

		structDefiners = append(structDefiners, definerFunc)
	}

	w.writeLinef("type %sVisitor interface {", mainClass)
	for _, method := range visitorMethods {
		w.writeLinef("  %s", method)
	}
	w.writeLinef("}")
	w.writeNewLine()

	print(structDefiners)
	for _, definer := range structDefiners {
		definer()
	}

	w.writeNewLine()
}

func main() {
	fmt.Println("here")

	args := os.Args[1:]
	fmt.Println(args)
	var dirName string
	if len(args) > 2 {
		fmt.Println("Usage: generate_expr -- <out dir>")
		os.Exit(1)
	}

	dirName = args[len(args)-1]

	asts := []struct {
		fileName  string
		mainClass string
		grammar   []string
	}{
		{"expr.go", "Expr",
			[]string{
				"Binary : Left Expr, Operator Token, Right Expr",
				"Logical: Left Expr, Operator Token, Right Expr",
				"Grouping : Expression Expr",
				"Literal : Value interface{}",
				"Unary : Operator Token, Right Expr",
				"TernaryCondition : Condition Expr, TrueBranch Expr, FalseBranch Expr",
				"Assign: Name Token, Value Expr",
				"Variable : Name Token",
				"Call : Callee Expr, Paren Token, Arguments []Expr",
				"Super : Super Token, Call Token",
				"Get : Object Expr, Name Token",
				"Set : Object Expr, Name Token, Value Expr",
				"Lambda : Name Token, Params []Token, Body []Stmt",
			}},
		{"stmt.go", "Stmt", []string{
			"IfStmt : Condition Expr, ThenBranch Stmt, ElseBranch Stmt",
			"WhileStmt : Condition Expr, Body Stmt",
			"ExprStmt: Expression Expr",
			"PrintStmt : Expression Expr",
			"VarStmt : Name Token, Initializer Expr",
			"FunctionStmt : Name Token, Params []Token, Body []Stmt",
			"ClassStmt : Name Token, SuperClass *Variable, Methods []*FunctionStmt",
			"BlockStmt : Statements []Stmt",
			"ReturnStmt : Keyword Token, Expression Expr",
		}},
	}

	for _, ast := range asts {
		fileName := path.Join(dirName, ast.fileName)

		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("Could not open file '%s' for writing: %v", fileName, err)
			os.Exit(1)
		}

		writer := bufio.NewWriter(file)
		generate(writer, ast.mainClass, ast.grammar)

		if err := writer.Flush(); err != nil {
			fmt.Printf("Could not write to file '%s': %v", fileName, err)
			os.Exit(1)
		}

		if err := file.Close(); err != nil {
			fmt.Printf("Could not close file '%s': %v", fileName, err)
			os.Exit(1)
		}
	}

}
