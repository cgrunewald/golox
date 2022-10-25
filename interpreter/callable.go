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
