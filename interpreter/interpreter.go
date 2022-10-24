package interpreter

import "fmt"

type InterpreterConfig struct {
	PrintFunc func(string)
}

var DefaultInterpreterConfig = InterpreterConfig{
	PrintFunc: func(s string) {
		fmt.Println(s)
	},
}

type Interpreter struct {
	config      InterpreterConfig
	environment *Environment
}

type result struct {
	Value interface{}
	Err   error
}

func Result(value interface{}) *result {
	return &result{value, nil}
}

func Error(err error) *result {
	return &result{nil, err}
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
	return &Interpreter{environment: NewEnvironment(), config: config}
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

	err := i.environment.Set(expr.Name, value.Value)
	if err != nil {
		return Error(err)
	}

	return value
}

func (i *Interpreter) VisitVariable(expr *Variable) interface{} {
	value, err := i.environment.Get(expr.Name)
	if err != nil {
		return Error(err)
	}
	return Result(value)
}

func (i *Interpreter) VisitLogical(expr *Logical) interface{} {
	return nil
}

func (i *Interpreter) VisitIfStmt(expr *IfStmt) interface{} {
	return nil
}

func (i *Interpreter) VisitWhileStmt(expr *WhileStmt) interface{} {
	return nil
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

	i.environment.Define(stmt.Name, value.Value)
	return Void
}

func (i *Interpreter) VisitBlockStmt(stmt *BlockStmt) interface{} {
	r := i.executeBlock(stmt.Statements, NewEnclosedEnvironment(i.environment))
	if r.(*result).IsError() {
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
		if r.(*result).IsError() {
			return r
		}
	}

	return Void
}

func (i *Interpreter) error(errType int32, token Token, message string) *result {
	err := NewRuntimeError(errType, token.Line, token.Lexeme, message)
	return Error(err)
}
