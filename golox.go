package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	i "github.com/cgrunewald/golox/interpreter"
)

var interpreter = i.NewInterpreter(i.InterpreterConfig{
	PrintFunc: func(value string) {
		fmt.Println(value)
	},
})

var resolver = i.NewResolver(interpreter)

func main() {
	args := os.Args[1:]
	if len(args) > 1 {
		fmt.Println("Usage: golox [script]")
		os.Exit(1)
		return
	} else if len(args) == 1 {
		err := runFile(args[0])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else {
		err := runPrompt()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
}

func runFile(file string) error {
	contents, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("golox: could not read file: '%s'", file)
	}

	str := string(contents)
	return run(str, false)
}

func runPrompt() error {
	for {
		fmt.Print("> ")

		reader := bufio.NewReader(os.Stdin)
		str, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		err = run(str, true)
		if err != nil && !i.IsLoxError(err) {
			return err
		} else if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func isTokenTypeStmt(tokenType i.TokenType) bool {
	return tokenType == i.TK_RIGHT_BRACE || tokenType == i.TK_SEMICOLON
}

func run(source string, interactive bool) error {
	scanner := i.NewScanner(source)
	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		return scanner.Errors()[0]
	}

	parser := i.NewParser(tokens)
	if interactive && len(tokens) > 1 && !isTokenTypeStmt(tokens[len(tokens)-2].TokenType) {
		expr := parser.ParseExpr()
		if parser.HasError() {
			return parser.Errors()[0]
		}

		resolver.ResolveExpr(expr)
		if resolver.HasError() {
			return resolver.Errors()[0]
		}

		result, err := interpreter.InterpretExpr(expr)
		if err != nil {
			return err
		}

		fmt.Println(result)
		return nil
	}

	program := parser.Parse()
	if parser.HasError() {
		return parser.Errors()[0]
	}

	resolver.ResolveStmts(program)
	if resolver.HasError() {
		return resolver.Errors()[0]
	}

	_, err := interpreter.Interpret(program)
	if err != nil {
		return err
	}

	return nil
}
