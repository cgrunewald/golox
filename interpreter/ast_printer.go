package interpreter

import (
	"fmt"
	"strings"
)

type ASTPrinter struct {
}

func NewASTPrinter() *ASTPrinter {
	return &ASTPrinter{}
}

func (p *ASTPrinter) VisitBinary(expr *Binary) interface{} {
	return p.parenthesized(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p *ASTPrinter) VisitGrouping(expr *Grouping) interface{} {
	return p.parenthesized("group", expr.Expression)
}

func (p *ASTPrinter) VisitLiteral(expr *Literal) interface{} {
	if str, ok := expr.Value.(string); ok {
		return fmt.Sprintf("\"%s\"", str)
	}
	return fmt.Sprintf("%v", expr.Value)
}

func (p *ASTPrinter) VisitUnary(expr *Unary) interface{} {
	return p.parenthesized(expr.Operator.Lexeme, expr.Right)
}

func (p *ASTPrinter) VisitTernaryCondition(expr *TernaryCondition) interface{} {
	return p.parenthesized("?:", expr.Condition, expr.TrueBranch, expr.FalseBranch)
}

func (p *ASTPrinter) VisitVariable(expr *Variable) interface{} {
	return p.variable(expr.Name)
}

func (p *ASTPrinter) VisitLogical(expr *Logical) interface{} {
	return p.parenthesized(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p *ASTPrinter) variable(name Token) string {
	return fmt.Sprintf("(var %s)", name.Lexeme)
}

func (p *ASTPrinter) VisitVarStmt(stmt *VarStmt) interface{} {
	return p.parenthesized("def "+stmt.Name.Lexeme, stmt.Initializer)
}

func (p *ASTPrinter) VisitFunctionStmt(stmt *FunctionStmt) interface{} {
	return p.printFunction(stmt.Name.Lexeme, stmt.Params, stmt.Body)
}

func (p *ASTPrinter) VisitLambda(expr *Lambda) interface{} {
	return p.printFunction("", expr.Params, expr.Body)
}

func (p *ASTPrinter) printFunction(name string, params []Token, body []Stmt) interface{} {
	builder := strings.Builder{}
	builder.WriteString("(def " + name)
	builder.WriteString("(")

	for i, param := range params {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(param.Lexeme)
	}

	builder.WriteString(") ")
	builder.WriteString(p.printStatements(body))
	builder.WriteString(")")

	return builder.String()
}

func (p *ASTPrinter) VisitExprStmt(stmt *ExprStmt) interface{} {
	return stmt.Expression.Accept(p)
}

func (p *ASTPrinter) VisitPrintStmt(stmt *PrintStmt) interface{} {
	return p.parenthesized("print", stmt.Expression)
}

func (p *ASTPrinter) VisitBlockStmt(stmt *BlockStmt) interface{} {
	return p.printStatements(stmt.Statements)
}

func (p *ASTPrinter) VisitAssign(expr *Assign) interface{} {
	return p.parenthesized("= "+p.variable(expr.Name), expr.Value)
}

func (p *ASTPrinter) VisitGet(expr *Get) interface{} {
	return p.parenthesized(fmt.Sprintf("get %q", expr.Name.Lexeme), expr.Object)
}

func (p *ASTPrinter) VisitSet(expr *Set) interface{} {
	return p.parenthesized(fmt.Sprintf("set %q", expr.Name.Lexeme), expr.Object, expr.Value)
}

func (p *ASTPrinter) VisitIfStmt(stmt *IfStmt) interface{} {
	expression := stmt.Condition.Accept(p)
	thenBranch := stmt.ThenBranch.Accept(p)
	var elseBranch interface{}
	if stmt.ElseBranch != nil {
		elseBranch = stmt.ElseBranch.Accept(p)
		return fmt.Sprintf("(if %s %s %s)", expression, thenBranch, elseBranch)
	}

	return fmt.Sprintf("(if %s %s)", expression, thenBranch)
}

func (p *ASTPrinter) VisitWhileStmt(stmt *WhileStmt) interface{} {
	expression := stmt.Condition.Accept(p)
	statement := stmt.Body.Accept(p)

	return fmt.Sprintf("(while %s %s)", expression, statement)
}

func (p *ASTPrinter) VisitCall(expr *Call) interface{} {
	builder := strings.Builder{}
	builder.WriteString("(call ")
	builder.WriteString(expr.Callee.Accept(p).(string))
	builder.WriteString(" ")
	builder.WriteString(p.parenthesized("arg", expr.Arguments...))
	builder.WriteString(")")
	return builder.String()
}

func (p *ASTPrinter) VisitReturnStmt(stmt *ReturnStmt) interface{} {
	return p.parenthesized("return", stmt.Expression)
}

func (p *ASTPrinter) printStatements(stmts []Stmt) string {
	builder := strings.Builder{}
	builder.WriteString("(scope")

	for _, stmt := range stmts {
		builder.WriteString(" ")
		builder.WriteString(stmt.Accept(p).(string))
	}

	builder.WriteString(")")
	return builder.String()
}

func (p *ASTPrinter) parenthesized(name string, exprs ...Expr) string {
	builder := strings.Builder{}

	builder.WriteString("(")
	builder.WriteString(name)

	for _, expr := range exprs {
		if expr != nil {
			builder.WriteString(" ")
			builder.WriteString(expr.Accept(p).(string))
		}
	}

	builder.WriteString(")")

	return builder.String()
}

func (p *ASTPrinter) VisitClassStmt(stmt *ClassStmt) interface{} {
	return ("class " + stmt.Name.Lexeme + ")")
}

func (p *ASTPrinter) PrintProgram(stmts []Stmt) string {
	return p.printStatements(stmts)
}

func (p *ASTPrinter) Print(expr Expr) string {
	str := expr.Accept(p)
	if str, ok := str.(string); ok {
		return str
	}

	return ""
}
