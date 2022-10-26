package interpreter

import "time"

type Callable interface {
	Arity() int
	Call(i *Interpreter, arguments []interface{}) interface{}
	String() string
}

type NativeCallable struct {
	arity    int
	callFunc func(i *Interpreter, arguments []interface{}) interface{}
}

func NewNativeCallable(arity int, callFunc func(i *Interpreter, arguments []interface{}) interface{}) Callable {
	return &NativeCallable{arity, callFunc}
}

func (n *NativeCallable) Arity() int {
	return n.arity
}

func (n *NativeCallable) Call(i *Interpreter, arguments []interface{}) interface{} {
	return n.callFunc(i, arguments)
}

func (n *NativeCallable) String() string {
	return "<native func>"
}

var ClockFunc = NewNativeCallable(0, func(i *Interpreter, arguments []interface{}) interface{} {
	return float64(time.Now().UnixMilli() / 1000.0)
})

type FunctionCallable struct {
	stmt               *FunctionStmt
	lexicalEnvironment *Environment
}

func NewFunctionCallable(stmt *FunctionStmt, lexicalEnvironment *Environment) Callable {
	return &FunctionCallable{stmt, lexicalEnvironment}
}

func (n *FunctionCallable) Arity() int {
	return len(n.stmt.Params)
}

func (n *FunctionCallable) Call(i *Interpreter, arguments []interface{}) interface{} {
	environment := NewEnclosedEnvironment(n.lexicalEnvironment)

	if len(arguments) != len(n.stmt.Params) {
		return i.error(E_INVALID_ARGUMENTS, n.stmt.Name, "Invalid arguments to function")
	}

	for idx, tok := range n.stmt.Params {
		environment.Define(tok.Lexeme, arguments[idx])
	}

	i.PushCallstack(n.String())
	defer i.PopCallstack()

	r := i.executeBlock(n.stmt.Body, environment).(*result)
	if r.IsError() {
		return r.Err
	}

	if r.IsStmtReturn {
		return r.Value
	}

	return nil
}

func (n *FunctionCallable) String() string {
	return n.stmt.Name.Lexeme
}
