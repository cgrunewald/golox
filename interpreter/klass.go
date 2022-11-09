package interpreter

import "fmt"

type Klass struct {
	name    Token
	methods map[string]*FunctionStmt
	env     *Environment
	super   *Klass
}

func NewKlass(name Token, methods []*FunctionStmt, env *Environment, super *Klass) *Klass {
	methodMap := make(map[string]*FunctionStmt)
	for _, method := range methods {
		methodMap[method.Name.Lexeme] = method
	}

	return &Klass{name: name, methods: methodMap, env: env, super: super}
}

func (k *Klass) Arity() int {
	if m, ok := k.methods["init"]; ok {
		return len(m.Params)
	}

	return 0
}

func (k *Klass) Call(i *Interpreter, arguments []interface{}) interface{} {
	instance := NewInstance(k)
	if _, ok := k.methods["init"]; ok {
		initMethod, _ := instance.Get("init")
		initMethod.(Callable).Call(i, arguments)
	}

	return instance
}

func (k *Klass) String() string {
	return k.name.Lexeme
}

type KlassInstance struct {
	klass      *Klass
	properties map[string]interface{}
}

type SuperReference struct {
	klass    *Klass
	instance *KlassInstance
}

func (r *SuperReference) Get(property string) (interface{}, bool) {
	klass, method := r.instance.FindMethod(property, r.klass)
	if method == nil {
		return nil, false
	}

	return r.instance.bind(property, method, klass), true
}

type Gettable interface {
	Get(property string) (interface{}, bool)
}

func NewInstance(klass *Klass) *KlassInstance {
	return &KlassInstance{klass, make(map[string]interface{})}
}

func (i *KlassInstance) String() string {
	return fmt.Sprintf("%v instance", i.klass)
}

func (i *KlassInstance) FindMethod(property string, klass *Klass) (*Klass, *FunctionStmt) {
	for klass != nil {
		method, ok := klass.methods[property]
		if ok {
			return klass, method
		}

		klass = klass.super
	}

	return nil, nil
}

func (i *KlassInstance) Get(property string) (interface{}, bool) {
	val, ok := i.properties[property]
	if ok {
		return val, ok
	}

	klass, method := i.FindMethod(property, i.klass)
	if method == nil {
		return nil, false
	}

	// Bind and cache the binding
	boundMethod := i.bind(property, method, klass)
	return boundMethod, true
}

func (i *KlassInstance) Set(property string, value interface{}) {
	i.properties[property] = value
}

func (i *KlassInstance) bind(property string, f *FunctionStmt, klass *Klass) Callable {
	methodEnv := NewEnclosedEnvironment(i.klass.env)
	methodEnv.Define("this", i)
	if klass.super != nil {
		methodEnv.Define("super", &SuperReference{klass: klass.super, instance: i})
	}

	if property == "init" {
		return NewInitFunctionCallable(f, methodEnv)
	}
	return NewFunctionCallable(f, methodEnv)
}
