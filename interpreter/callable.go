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
	name               Token
	params             []Token
	body               []Stmt
	lexicalEnvironment *Environment
	isInit             bool
}

func NewFunctionCallable(stmt *FunctionStmt, lexicalEnvironment *Environment) Callable {
	return &FunctionCallable{name: stmt.Name, params: stmt.Params, body: stmt.Body, lexicalEnvironment: lexicalEnvironment, isInit: false}
}

func NewInitFunctionCallable(stmt *FunctionStmt, lexicalEnvironment *Environment) Callable {
	return &FunctionCallable{name: stmt.Name, params: stmt.Params, body: stmt.Body, lexicalEnvironment: lexicalEnvironment, isInit: true}
}

func NewLambdaCallable(expr *Lambda, lexicalEnvironment *Environment) Callable {
	return &FunctionCallable{name: expr.Name, params: expr.Params, body: expr.Body, lexicalEnvironment: lexicalEnvironment}
}

func (n *FunctionCallable) Arity() int {
	return len(n.params)
}

func (n *FunctionCallable) Call(i *Interpreter, arguments []interface{}) interface{} {
	environment := NewEnclosedEnvironment(n.lexicalEnvironment)

	if len(arguments) != len(n.params) {
		return i.error(E_INVALID_ARGUMENTS, n.name, "Invalid arguments to function")
	}

	for idx, tok := range n.params {
		environment.Define(tok.Lexeme, arguments[idx])
	}

	i.PushCallstack(n.String())
	defer i.PopCallstack()

	r := i.executeBlock(n.body, environment).(*result)
	if r.IsError() {
		return r.Err
	}

	if n.isInit {
		val, err := environment.GetAt(ThisToken, 0)
		if err != nil {
			return err
		}
		return val
	}

	if r.IsStmtReturn {
		return r.Value
	}

	return nil
}

func (n *FunctionCallable) String() string {
	return n.name.Lexeme
}
