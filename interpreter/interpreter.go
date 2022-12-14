package interpreter

import "fmt"

type InterpreterConfig struct {
	PrintFunc           func(string)
	GlobalFuncOverrides map[string]Callable
}

var DefaultInterpreterConfig = InterpreterConfig{
	PrintFunc: func(s string) {
		fmt.Println(s)
	},
	GlobalFuncOverrides: nil,
}

type Interpreter struct {
	config            InterpreterConfig
	globalEnvironment *Environment
	environment       *Environment
	callstack         []string
	locals            map[Expr]int
}

type result struct {
	Value        interface{}
	IsStmtReturn bool
	Err          error
}

func Result(value interface{}) *result {
	return &result{value, false, nil}
}

func Error(err error) *result {
	return &result{nil, false, err}
}

func Return(value interface{}) *result {
	return &result{value, true, nil}
}

func (r *result) IsBlockBreaking() bool {
	return r.IsError() || r.IsStmtReturn
}

var Void = Result(nil)

func (r *result) IsError() bool {
	return r.Err != nil
}

func (r *result) ToNumber() (float64, bool) {
	if r.IsError() {
		return 0, false
	}
	value, ok := r.Value.(float64)
	return value, ok
}

func (r *result) coerceString() string {
	// Expect the caller to check for errors
	if r.IsError() {
		panic("Cannot coerce error")
	}
	return fmt.Sprintf("%v", r.Value)
}

func (r *result) ToString() (string, bool) {
	if r.IsError() {
		return "", false
	}
	value, ok := r.Value.(string)
	return value, ok
}

func (r *result) IsString() bool {
	if r.IsError() {
		return false
	}

	_, ok := r.Value.(string)
	return ok
}

func (r *result) IsNumber() bool {
	if r.IsError() {
		return false
	}

	_, ok := r.Value.(float64)
	return ok
}

func (r *result) IsTruthy() bool {
	if r.IsError() {
		return false
	}

	if r.Value == nil {
		return false
	}

	if b, ok := r.Value.(bool); ok {
		return b
	}

	return false
}

func NewInterpreter(config InterpreterConfig) *Interpreter {
	globals := NewEnvironment()
	globals.Define("clock", ClockFunc)

	if config.GlobalFuncOverrides != nil {
		for key, value := range config.GlobalFuncOverrides {
			globals.Define(key, value)
		}
	}

	return &Interpreter{environment: globals, config: config, globalEnvironment: globals, callstack: make([]string, 0), locals: make(map[Expr]int)}
}

func (i *Interpreter) resolve(expr Expr, hops int) {
	i.locals[expr] = hops
}

func (i *Interpreter) PushCallstack(functionName string) {
	i.callstack = append(i.callstack, functionName)
}

func (i *Interpreter) PopCallstack() {
	i.callstack = i.callstack[:len(i.callstack)-1]
}

func (i *Interpreter) Interpret(stmt []Stmt) (interface{}, error) {
	r := i.executeGlobalBlock(stmt)
	if re, ok := r.(*result); ok {
		if re.IsError() {
			return nil, re.Err
		}
	}

	return r, nil
}

func (i *Interpreter) InterpretExpr(expr Expr) (interface{}, error) {
	result := expr.Accept(i).(*result)
	if result.IsError() {
		return nil, result.Err
	}
	return result.Value, nil
}

func (i *Interpreter) evaluateExpression(expr Expr) *result {
	r := expr.Accept(i).(*result)
	return r
}

func (i *Interpreter) doArithmetic(expr *Binary, left *result, right *result, f func(l float64, r float64) float64) *result {
	l1, ok := left.ToNumber()
	if !ok {
		return i.error(E_UNEXPECTED_TYPE, expr.Operator, "Left operand must be a number.")
	}

	l2, ok := right.ToNumber()
	if !ok {
		return i.error(E_UNEXPECTED_TYPE, expr.Operator, "Right operand must be a number.")
	}

	return Result(f(l1, l2))
}

func (i *Interpreter) doStringComparison(expr *Binary, left *result, right *result, f func(l string, r string) bool) *result {
	l1, ok := left.ToString()
	if !ok {
		return i.error(E_UNEXPECTED_TYPE, expr.Operator, "Left operand must be a string.")
	}

	l2, ok := right.ToString()
	if !ok {
		return i.error(E_UNEXPECTED_TYPE, expr.Operator, "Right operand must be a string.")
	}

	return Result(f(l1, l2))
}

func (i *Interpreter) doNumComparison(expr *Binary, left *result, right *result, f func(l float64, r float64) bool) *result {
	l1, ok := left.ToNumber()
	if !ok {
		return i.error(E_UNEXPECTED_TYPE, expr.Operator, "Left operand must be a number.")
	}

	l2, ok := right.ToNumber()
	if !ok {
		return i.error(E_UNEXPECTED_TYPE, expr.Operator, "Right operand must be a number.")
	}

	return Result(f(l1, l2))
}

func (i *Interpreter) VisitBinary(expr *Binary) interface{} {
	left := expr.Left.Accept(i).(*result)
	if left.IsError() {
		return left
	}

	right := expr.Right.Accept(i).(*result)
	if right.IsError() {
		return right
	}

	switch expr.Operator.TokenType {
	case TK_PLUS:
		if left.IsString() || right.IsString() && (!left.IsError() && !right.IsError()) {
			sl := left.coerceString()
			sr := right.coerceString()

			return Result(sl + sr)
		}
		return i.doArithmetic(expr, left, right, func(l float64, r float64) float64 { return l + r })
	case TK_MINUS:
		return i.doArithmetic(expr, left, right, func(l float64, r float64) float64 { return l - r })
	case TK_STAR:
		return i.doArithmetic(expr, left, right, func(l float64, r float64) float64 { return l * r })
	case TK_SLASH:
		if num, ok := right.ToNumber(); ok && num == 0.0 {
			return i.error(E_DIVIDE_BY_ZERO, expr.Operator, "Cannot divide by zero.")
		}
		return i.doArithmetic(expr, left, right, func(l float64, r float64) float64 { return l / r })
	case TK_BANG_EQUAL:
		return Result(left.Value != right.Value)
	case TK_EQUAL_EQUAL:
		return Result(left.Value == right.Value)
	case TK_GREATER:
		if left.IsString() {
			return i.doStringComparison(expr, left, right, func(l string, r string) bool { return l > r })
		}
		return i.doNumComparison(expr, left, right, func(l float64, r float64) bool { return l > r })
	case TK_GREATER_EQUAL:
		if left.IsString() {
			return i.doStringComparison(expr, left, right, func(l string, r string) bool { return l >= r })
		}
		return i.doNumComparison(expr, left, right, func(l float64, r float64) bool { return l >= r })
	case TK_LESS:
		if left.IsString() {
			return i.doStringComparison(expr, left, right, func(l string, r string) bool { return l < r })
		}
		return i.doNumComparison(expr, left, right, func(l float64, r float64) bool { return l < r })
	case TK_LESS_EQUAL:
		if left.IsString() {
			return i.doStringComparison(expr, left, right, func(l string, r string) bool { return l <= r })
		}
		return i.doNumComparison(expr, left, right, func(l float64, r float64) bool { return l <= r })
	}

	return i.error(E_UNEXPECTED_OPERATOR, expr.Operator, "Invalid binary operator")
}

func (i *Interpreter) VisitGrouping(expr *Grouping) interface{} {
	return expr.Expression.Accept(i)
}

func (i *Interpreter) VisitLiteral(expr *Literal) interface{} {
	return Result(expr.Value)
}

func (i *Interpreter) VisitUnary(expr *Unary) interface{} {
	result := expr.Right.Accept(i).(*result)
	if result.IsError() {
		return result
	}

	if expr.Operator.TokenType == TK_MINUS {
		number, ok := result.ToNumber()
		if !ok {
			return i.error(E_UNEXPECTED_TYPE, expr.Operator, "Operand must be a number.")
		}
		return Result(-number)
	} else if expr.Operator.TokenType == TK_BANG {
		return Result(!result.IsTruthy())
	}

	// Should never get here
	return i.error(E_UNEXPECTED_OPERATOR, expr.Operator, "Invalid unary operator")
}

func (i *Interpreter) VisitTernaryCondition(expr *TernaryCondition) interface{} {
	truthCond := expr.Condition.Accept(i).(*result)
	if truthCond.IsError() {
		return truthCond
	}

	if truthCond.IsTruthy() {
		return expr.TrueBranch.Accept(i)
	} else {
		return expr.FalseBranch.Accept(i)
	}
}

func (i *Interpreter) VisitAssign(expr *Assign) interface{} {
	value := expr.Value.Accept(i).(*result)
	if value.IsError() {
		return value
	}

	distance, ok := i.locals[expr]
	if ok {
		err := i.environment.SetAt(expr.Name, value.Value, distance)
		if err != nil {
			return Error(err)
		}
	} else {
		err := i.globalEnvironment.Set(expr.Name, value.Value)
		if err != nil {
			return Error(err)
		}
	}

	return value
}

func (i *Interpreter) lookupVariable(name Token, expr Expr) interface{} {
	distance, ok := i.locals[expr]
	if ok {
		value, err := i.environment.GetAt(name, distance)
		if err != nil {
			return Error(err)
		}
		return Result(value)
	} else {
		value, err := i.globalEnvironment.Get(name)
		if err != nil {
			return Error(err)
		}
		return Result(value)
	}
}

func (i *Interpreter) VisitVariable(expr *Variable) interface{} {
	return i.lookupVariable(expr.Name, expr)
}

func (i *Interpreter) VisitLogical(expr *Logical) interface{} {
	left := expr.Left.Accept(i).(*result)
	if left.IsError() {
		return left
	}

	if expr.Operator.TokenType == TK_OR {
		if left.IsTruthy() {
			return Result(true)
		}

		right := expr.Right.Accept(i).(*result)
		if right.IsError() {
			return right
		}

		return Result(right.IsTruthy())
	} else if expr.Operator.TokenType == TK_AND {
		right := expr.Right.Accept(i).(*result)
		if right.IsError() {
			return right
		}

		return Result(left.IsTruthy() && right.IsTruthy())
	} else {
		return i.error(E_UNEXPECTED_OPERATOR, expr.Operator, fmt.Sprintf("unexpected operator %s", expr.Operator.Lexeme))
	}
}

func (i *Interpreter) VisitIfStmt(expr *IfStmt) interface{} {
	rCond := expr.Condition.Accept(i).(*result)
	if rCond.IsError() {
		return rCond
	}

	if rCond.IsTruthy() {
		return expr.ThenBranch.Accept(i)
	} else {
		if expr.ElseBranch != nil {
			return expr.ElseBranch.Accept(i)
		}
	}

	return Void
}

func (i *Interpreter) VisitWhileStmt(expr *WhileStmt) interface{} {
	for {
		rCond := expr.Condition.Accept(i).(*result)
		if rCond.IsError() {
			return rCond
		}

		if !rCond.IsTruthy() {
			break
		}

		rBody := expr.Body.Accept(i).(*result)
		if rBody.IsBlockBreaking() {
			return rBody
		}
	}

	return Void
}

func (i *Interpreter) VisitGet(expr *Get) interface{} {
	object := i.evaluateExpression(expr.Object)
	if object.IsError() {
		return object
	}

	instance, ok := object.Value.(Gettable)
	if !ok {
		return i.error(E_NOT_AN_OBJECT, expr.Name, "Expression does not evaluate to an object")
	}

	val, ok := instance.Get(expr.Name.Lexeme)
	if !ok {
		return i.error(E_UNDEFINED_OBJECT_PROPERTY, expr.Name, "Property is not defined on object")
	}

	return Result(val)
}

func (i *Interpreter) VisitSuper(expr *Super) interface{} {
	superVar := i.lookupVariable(expr.Super, expr).(*result)
	if superVar.IsError() {
		return superVar
	}

	klass, ok := superVar.Value.(*Klass)
	if !ok {
		return i.error(E_NOT_AN_OBJECT, expr.Super, "super variable is not an object")
	}

	distance := i.locals[expr]
	thisInstance, err := i.environment.GetAt(ThisToken, distance-1)
	if err != nil {
		return Error(err)
	}

	this, ok := thisInstance.(*KlassInstance)
	if !ok {
		panic("should never get here")
	}

	result, err := klass.GetSuperMethod(expr.Call, this)
	if err != nil {
		return Error(err)
	}
	return Result(result)
}

func (i *Interpreter) VisitSet(expr *Set) interface{} {
	object := i.evaluateExpression(expr.Object)
	if object.IsError() {
		return object
	}

	value := i.evaluateExpression(expr.Value)
	if value.IsError() {
		return object
	}

	instance, ok := object.Value.(*KlassInstance)
	if !ok {
		return i.error(E_UNDEFINED_OBJECT_PROPERTY, expr.Name, "Property is not defined on object")
	}

	instance.Set(expr.Name.Lexeme, value.Value)

	return value
}

func (i *Interpreter) VisitCall(expr *Call) interface{} {
	rCallee := expr.Callee.Accept(i).(*result)
	if rCallee.IsError() {
		return rCallee
	}

	callable, ok := rCallee.Value.(Callable)
	if !ok {
		return i.error(E_CANNOT_CALL, expr.Paren, "Can only call functions or classes")
	}

	if callable.Arity() != len(expr.Arguments) {
		return i.error(E_INVALID_ARGUMENTS, expr.Paren, "Provided arguments do not match function definition")
	}

	argValues := make([]interface{}, 0)

	for _, value := range expr.Arguments {
		argValue := value.Accept(i).(*result)
		if argValue.IsError() {
			return argValue
		}

		argValues = append(argValues, argValue.Value)
	}

	callResult := callable.Call(i, argValues)
	if err, ok := callResult.(error); ok {
		return Error(err)
	}
	return Result(callResult)
}

func (i *Interpreter) VisitExprStmt(stmt *ExprStmt) interface{} {
	r := stmt.Expression.Accept(i)
	if r.(*result).IsError() {
		return r
	}

	return Void
}

func (i *Interpreter) VisitPrintStmt(stmt *PrintStmt) interface{} {
	value := stmt.Expression.Accept(i).(*result)
	if value.IsError() {
		return value
	}

	if i.config.PrintFunc != nil {
		i.config.PrintFunc(fmt.Sprintf("%v", value.Value))
	}

	return Void
}

func (i *Interpreter) VisitVarStmt(stmt *VarStmt) interface{} {
	value := Void
	if stmt.Initializer != nil {
		value = stmt.Initializer.Accept(i).(*result)
	}

	if value.IsError() {
		return value
	}

	i.environment.Define(stmt.Name.Lexeme, value.Value)
	return Void
}

func (i *Interpreter) VisitFunctionStmt(stmt *FunctionStmt) interface{} {
	callable := NewFunctionCallable(stmt, i.environment)
	i.environment.Define(stmt.Name.Lexeme, callable)
	return Void
}

func (i *Interpreter) VisitLambda(expr *Lambda) interface{} {
	return Result(NewLambdaCallable(expr, i.environment))
}

func (i *Interpreter) VisitBlockStmt(stmt *BlockStmt) interface{} {
	r := i.executeBlock(stmt.Statements, NewEnclosedEnvironment(i.environment))
	if r.(*result).IsBlockBreaking() {
		return r
	}
	return Void
}

func (i *Interpreter) executeGlobalBlock(stmts []Stmt) interface{} {
	return i.executeBlock(stmts, nil)
}

func (i *Interpreter) executeBlock(statements []Stmt, env *Environment) interface{} {
	if env != nil {
		previous := i.environment
		defer func() { i.environment = previous }()

		i.environment = env
	}

	for _, stmt := range statements {
		r := stmt.Accept(i)
		if r.(*result).IsBlockBreaking() {
			return r
		}
	}

	return Void
}

func (i *Interpreter) VisitReturnStmt(stmt *ReturnStmt) interface{} {
	if len(i.callstack) == 0 {
		return i.error(E_UNEXPECTED_RETURN, stmt.Keyword, "unexpected return in current scope")
	}

	var returnValue interface{}
	if stmt.Expression != nil {
		stmtResult := stmt.Expression.Accept(i).(*result)
		if stmtResult.IsError() {
			return stmtResult
		}

		returnValue = stmtResult.Value
	}
	return Return(returnValue)
}

func (i *Interpreter) VisitClassStmt(stmt *ClassStmt) interface{} {
	var superKlass *Klass
	if stmt.SuperClass != nil {
		val := i.evaluateExpression(stmt.SuperClass)
		if val.IsError() {
			return val
		}

		sKlass, ok := val.Value.(*Klass)
		if !ok {
			return i.error(E_INVALID_CLASS, stmt.SuperClass.Name, "Invalid super class")
		}

		superKlass = sKlass
	}
	i.environment.Define(stmt.Name.Lexeme, NewKlass(stmt.Name, stmt.Methods, i.environment, superKlass))
	return Void
}

func (i *Interpreter) error(errType int32, token Token, message string) *result {
	err := NewRuntimeError(errType, token.Line, token.Lexeme, message)
	return Error(err)
}
