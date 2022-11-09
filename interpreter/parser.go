package interpreter

type Parser struct {
	tokens []Token
	errors []error

	current int
}

const MaxArguments = 255

/*

Grammar:

program        → declaration* EOF ;

declaration    → funDecl | varDecl | classDecl | statement ;

varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;
funDecl        → "fun" IDENTIFIER "(" parameters? ")" blockStmt ;
classDecl      → "class" IDENTIFIER ( "<" IDENTIFIER )? "{" ( varDecl | funDecl )* "}";

statement			 → exprStmt | printStmt | blockStmt | ifStmt | forStmt | whileStmt | returnStmt;
exprStmt       → expression ";" ;
printStmt      → "print" expression ";" ;
ifStmt				 → "if" "(" expression ")" statement ( "else" statement )? ;
whileStmt			 → "while" "(" expression ")" statement ;
forStmt				 → "for" "(" ( varDecl | exprStmt | ";" ) expression? ";" expression? ")" statement ;
returnStmt     → "return" ( expression? ) ";" ;

expression     → ternary ;
assignment 	   → ( call "." )? IDENTIFIER "=" assignment | ternary;
ternary				 → logical_or ( "?" expression ":" expression )? ;
logical_or		 → logical_and ( "or" logical_and )* ;
logical_and		 → equality ( "and" equality )* ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term 					 → factor ( ( "-" | "+" ) factor )* ;
factor  			 → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary | call ;
call           → primary ( ( "(" arguments? ")" ) | ( "." IDENTIFIER ) ) * ;
primary        → IDENTIFIER | NUMBER | STRING | "true" | "false" | "nil" | lambda | ( "(" expression ")" ) ;

lambda         → "fun" "(" parameters? ")" blockStmt ;
arguments      → expression ( "," expression )* ;
parameters     → IDENTIFIER ( "," IDENTIFIER )* ;

*/

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, errors: make([]error, 0)}
}

func (p *Parser) statement() (Stmt, error) {
	if p.match(TK_PRINT) {
		return p.printStmt()
	}

	if p.match(TK_LEFT_BRACE) {
		return p.blockStmt()
	}

	if p.match(TK_FOR) {
		return p.forStmt()
	}

	if p.match(TK_WHILE) {
		return p.whileStmt()
	}

	if p.match(TK_IF) {
		return p.ifStmt()
	}

	if p.match(TK_RETURN) {
		return p.returnStmt()
	}

	return p.exprStmt()
}

func (p *Parser) returnStmt() (Stmt, error) {
	retToken := p.previous()

	var expression Expr
	var err error
	if !p.check(TK_SEMICOLON) {
		expression, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(TK_SEMICOLON, "expected semicolon after return")
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{Keyword: retToken, Expression: expression}, nil
}

func (p *Parser) ifStmt() (Stmt, error) {
	_, errPL := p.consume(TK_LEFT_PAREN, "expected left parenthesis")
	if errPL != nil {
		return nil, errPL
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, errPR := p.consume(TK_RIGHT_PAREN, "expected right parenthesis")
	if errPR != nil {
		return nil, errPR
	}

	thenBranch, errThen := p.statement()
	if errThen != nil {
		return nil, errThen
	}

	var elseBranch Stmt
	var errElse error
	if p.match(TK_ELSE) {
		elseBranch, errElse = p.statement()
		if errElse != nil {
			return nil, errElse
		}
	}

	return &IfStmt{Condition: expr, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}

func (p *Parser) whileStmt() (Stmt, error) {
	_, errPL := p.consume(TK_LEFT_PAREN, "expected left parenthesis")
	if errPL != nil {
		return nil, errPL
	}

	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, errPR := p.consume(TK_RIGHT_PAREN, "expected right parenthesis")
	if errPR != nil {
		return nil, errPR
	}

	stmt, errStmt := p.statement()
	if errStmt != nil {
		return nil, errStmt
	}

	return &WhileStmt{Condition: expr, Body: stmt}, nil
}

func (p *Parser) forStmt() (Stmt, error) {
	_, errPL := p.consume(TK_LEFT_PAREN, "expected left parenthesis")
	if errPL != nil {
		return nil, errPL
	}

	var initStmt Stmt
	var initErr error
	if !p.match(TK_SEMICOLON) {
		if p.match(TK_VAR) {
			initStmt, initErr = p.varDecl()
		} else {
			initStmt, initErr = p.exprStmt()
		}
	}

	if initErr != nil {
		return nil, initErr
	}

	var condExpr Expr
	var condErr error
	if !p.check(TK_SEMICOLON) {
		condExpr, condErr = p.expression()
	}

	if condErr != nil {
		return nil, condErr
	}
	p.consume(TK_SEMICOLON, "semicolon expected after for-loop condition")

	var incrExpr Expr
	var incrErr error
	if !p.check(TK_RIGHT_PAREN) {
		incrExpr, incrErr = p.expression()
	}

	if incrErr != nil {
		return nil, incrErr
	}

	p.consume(TK_RIGHT_PAREN, "right parenthesis expected to close for-loop")

	stmt, err := p.statement()
	if err != nil {
		return nil, err
	}

	if incrExpr != nil {
		stmt = &BlockStmt{Statements: []Stmt{stmt, &ExprStmt{Expression: incrExpr}}}
	}

	if condExpr == nil {
		condExpr = &Literal{Value: true}
	}

	stmt = &WhileStmt{Condition: condExpr, Body: stmt}

	if initStmt != nil {
		stmt = &BlockStmt{Statements: []Stmt{initStmt, stmt}}
	}

	return stmt, nil
}

func (p *Parser) exprStmt() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	p.consume(TK_SEMICOLON, "Expect ';' after value.")
	return &ExprStmt{Expression: expr}, nil
}

func (p *Parser) blockStmt() (Stmt, error) {
	stmts := make([]Stmt, 0)
	for !p.isAtEnd() && p.peek().TokenType != TK_RIGHT_BRACE {
		stmt, err := p.declaration()
		if err != nil {
			p.synchronize()
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	p.consume(TK_RIGHT_BRACE, "Expect '}' after block.")

	return &BlockStmt{Statements: stmts}, nil
}

func (p *Parser) printStmt() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	p.consume(TK_SEMICOLON, "Expect ';' after value.")
	return &PrintStmt{Expression: expr}, nil
}

func (p *Parser) declaration() (Stmt, error) {
	if p.match(TK_VAR) {
		return p.varDecl()
	}

	if p.check(TK_FUN) && p.checkNext(TK_IDENTIFIER) {
		p.consume(TK_FUN, "")
		return p.functionDecl()
	}

	if p.match(TK_CLASS) {
		return p.classDecl()
	}

	return p.statement()
}

func (p *Parser) lambda() (Expr, error) {
	token := p.previous()

	stmt, err := p.finishFunction(token)
	if err != nil {
		return nil, err
	}

	fStmt := stmt.(*FunctionStmt)
	return &Lambda{Name: fStmt.Name, Params: fStmt.Params, Body: fStmt.Body}, nil
}

func (p *Parser) finishFunction(token Token) (Stmt, error) {
	var err error
	_, err = p.consume(TK_LEFT_PAREN, "Expected '(' for parameter list")
	if err != nil {
		return nil, err
	}

	tokList := make([]Token, 0)
	if p.check(TK_RIGHT_PAREN) {
		goto final
	}

	for {
		paramTok, paramErr := p.consume(TK_IDENTIFIER, "Expected identifier for parameter")
		if paramErr != nil {
			return nil, paramErr
		}

		tokList = append(tokList, paramTok)
		if !p.match(TK_COMMA) {
			break
		}
	}

final:
	_, err = p.consume(TK_RIGHT_PAREN, "Expected ')' to end parameter list")
	if err != nil {
		return nil, err
	}

	_, err = p.consume(TK_LEFT_BRACE, "Expected '{' to begin function body")
	if err != nil {
		return nil, err
	}

	block, err := p.blockStmt()
	if err != nil {
		return nil, err
	}

	return &FunctionStmt{Name: token, Params: tokList, Body: block.(*BlockStmt).Statements}, nil
}

func (p *Parser) classDecl() (Stmt, error) {
	idToken, err := p.consume(TK_IDENTIFIER, "Expected class name")
	if err != nil {
		return nil, err
	}

	var superToken *Token
	if p.match(TK_LESS) {
		tok, err := p.consume(TK_IDENTIFIER, "Expected superclass name")
		if err != nil {
			return nil, err
		}

		superToken = &tok
	}

	_, err = p.consume(TK_LEFT_BRACE, "Expected '{' to open class definition")
	if err != nil {
		return nil, err
	}

	functions := make([]*FunctionStmt, 0)

	for p.match(TK_FUN) {
		fStmt, err := p.functionDecl()
		if err != nil {
			return nil, err
		}

		functions = append(functions, fStmt.(*FunctionStmt))
	}

	_, err = p.consume(TK_RIGHT_BRACE, "Expected '}' to close class definition")
	if err != nil {
		return nil, err
	}

	return &ClassStmt{Name: idToken, Methods: functions, SuperClass: superToken}, nil
}

func (p *Parser) functionDecl() (Stmt, error) {
	idToken, err := p.consume(TK_IDENTIFIER, "Expected function name")
	if err != nil {
		return nil, err
	}

	return p.finishFunction(idToken)
}

func (p *Parser) varDecl() (Stmt, error) {
	token, err := p.consume(TK_IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var initializer Expr
	if p.match(TK_EQUAL) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	p.consume(TK_SEMICOLON, "Expect ';' after variable declaration.")
	return &VarStmt{Name: token, Initializer: initializer}, nil
}

func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (Expr, error) {
	expr, err := p.ternary()
	if err != nil {
		return nil, err
	}

	if p.match(TK_EQUAL) {
		equals := p.previous()
		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		if variable, ok := expr.(*Variable); ok {
			return &Assign{Name: variable.Name, Value: value}, nil
		} else if get, ok := expr.(*Get); ok {
			return &Set{Object: get.Object, Name: get.Name, Value: value}, nil
		}

		return nil, p.error(equals, "Invalid assignment target.")
	}

	return expr, nil
}

func (p *Parser) ternary() (Expr, error) {
	expr, err := p.logicalOr()
	if err != nil {
		return nil, err
	}

	if p.match(TK_QUESTION) {
		trueExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		p.consume(TK_COLON, "Expect ':' after true expression.")

		falseExpr, err := p.expression()
		if err != nil {
			return nil, err
		}

		return &TernaryCondition{Condition: expr, TrueBranch: trueExpr, FalseBranch: falseExpr}, nil
	}

	return expr, nil
}

func (p *Parser) logicalOr() (Expr, error) {
	expr, err := p.logicalAnd()
	if err != nil {
		return nil, err
	}

	for p.match(TK_OR) {
		op := p.previous()
		right, err := p.logicalAnd()
		if err != nil {
			return nil, err
		}
		expr = &Logical{Left: expr, Operator: op, Right: right}
	}

	return expr, nil
}

func (p *Parser) logicalAnd() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(TK_AND) {
		op := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = &Logical{Left: expr, Operator: op, Right: right}
	}

	return expr, nil
}

func (p *Parser) equality() (Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(TK_BANG_EQUAL, TK_EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) comparison() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(TK_LESS, TK_LESS_EQUAL, TK_GREATER, TK_GREATER_EQUAL) {
		operator := p.previous()

		right, err := p.term()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(TK_PLUS, TK_MINUS) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) factor() (Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(TK_STAR, TK_SLASH) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = &Binary{Left: expr, Operator: operator, Right: right}
	}

	return expr, nil
}

func (p *Parser) unary() (Expr, error) {
	if p.match(TK_BANG, TK_MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return &Unary{Operator: operator, Right: right}, nil
	}

	return p.call()
}

func (p *Parser) call() (Expr, error) {
	primary, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(TK_LEFT_PAREN) {
			primary, err = p.finishCall(primary)
			if err != nil {
				return nil, err
			}
		} else if p.match(TK_DOT) {
			identifier, err := p.consume(TK_IDENTIFIER, "Expected identifier in a dot expression")
			if err != nil {
				return nil, err
			}

			primary = &Get{Object: primary, Name: identifier}
		} else {
			break
		}
	}

	return primary, nil
}

func (p *Parser) finishCall(expr Expr) (Expr, error) {
	exprList := make([]Expr, 0)

	if p.check(TK_RIGHT_PAREN) {
		goto finish
	}

	for {
		argExpr, argErr := p.expression()
		if argErr != nil {
			return nil, argErr
		}

		exprList = append(exprList, argExpr)

		if !p.match(TK_COMMA) {
			break
		}
	}

finish:
	p.consume(TK_RIGHT_PAREN, "expected ')' after arguments")
	if len(exprList) > MaxArguments {
		// Note that we continue parsing
		p.error(p.previous(), "Argument list exceeded maximum length")
	}
	return &Call{Callee: expr, Arguments: exprList, Paren: p.previous()}, nil
}

func (p *Parser) primary() (Expr, error) {
	if p.match(TK_FALSE) {
		return &Literal{Value: false}, nil
	} else if p.match(TK_TRUE) {
		return &Literal{Value: true}, nil
	} else if p.match(TK_NIL) {
		return &Literal{Value: nil}, nil
	} else if p.match(TK_NUMBER, TK_STRING) {
		return &Literal{Value: p.previous().Literal}, nil
	} else if p.match(TK_IDENTIFIER) {
		return &Variable{Name: p.previous()}, nil
	} else if p.match(TK_THIS) {
		return &Variable{Name: p.previous()}, nil
	} else if p.match(TK_SUPER) {
		return &Variable{Name: p.previous()}, nil
	} else if p.match(TK_FUN) {
		return p.lambda()
	} else if p.match(TK_LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		p.consume(TK_RIGHT_PAREN, "Expect ')' after expression.")
		return &Grouping{Expression: expr}, nil
	}

	return nil, p.error(p.peek(), "Expected expression.")
}

func (p *Parser) error(t Token, msg string) error {
	var err error
	if t.TokenType == TK_EOF {
		err = NewTokenError(t.Line, " at end", msg)
	} else {
		err = NewTokenError(t.Line, " at '"+t.Lexeme+"'", msg)
	}

	p.errors = append(p.errors, err)
	return err
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().TokenType == t
}

func (p *Parser) checkNext(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.tokens[p.current+1].TokenType == t
}

func (p *Parser) consume(t TokenType, msg string) (Token, error) {
	if p.check(t) {
		return p.advance(), nil
	}

	return Token{}, p.error(p.peek(), msg)
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == TK_EOF
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) synchronize() {
	for !p.isAtEnd() {
		token := p.advance()
		if token.TokenType == TK_SEMICOLON {
			return
		}
	}
}

func (p *Parser) ParseExpr() Expr {
	// Reset the errors
	p.errors = make([]error, 0)
	p.current = 0

	expr, _ := p.expression()
	return expr
}

func (p *Parser) Parse() []Stmt {
	// Reset the errors
	p.errors = make([]error, 0)
	p.current = 0

	stmts := make([]Stmt, 0)

	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			p.synchronize()
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	return stmts
}

func (p *Parser) HasError() bool {
	return len(p.errors) > 0
}

func (p *Parser) Errors() []error {
	return p.errors
}
