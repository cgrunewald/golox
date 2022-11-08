package interpreter

import (
	"errors"
	"fmt"
)

const (
	E_NO_ERROR int32 = iota
	E_UNEXPECTED_TYPE
	E_UNEXPECTED_OPERATOR
	E_UNDEFINED_VARIABLE
	E_CANNOT_CALL
	E_INVALID_ARGUMENTS
	E_DIVIDE_BY_ZERO
	E_UNEXPECTED_RETURN
	E_VAR_ALREADY_DEFINED
	E_NOT_AN_OBJECT
	E_UNDEFINED_OBJECT_PROPERTY
)

type LoxError struct {
	line             int
	message          string
	where            string
	runtimeErrorType int32
}

func (err *LoxError) Error() string {
	return fmt.Sprintf("[line %d] Error%s: %s\n", err.line, err.where, err.message)
}

func NewError(line int, message string) error {
	return &LoxError{line: line, message: message, where: "", runtimeErrorType: E_NO_ERROR}
}

func NewRuntimeError(errorType int32, line int, where string, message string) error {
	return &LoxError{runtimeErrorType: errorType, line: line, message: message, where: where}
}

func NewTokenError(line int, where string, message string) error {
	return &LoxError{line: line, message: message, where: where, runtimeErrorType: E_NO_ERROR}
}

func IfLoxError(err error, callback func(*LoxError)) bool {
	if err == nil {
		return false
	}

	var loxError *LoxError
	if errors.As(err, &loxError) {
		callback(loxError)
		return true
	}
	return false
}

func IsLoxError(err error) bool {
	if err == nil {
		return false
	}

	var loxError *LoxError
	return errors.As(err, &loxError)
}

func (t Token) ToError(msg string) error {
	return NewTokenError(t.Line, t.Lexeme, msg)
}

func (t Token) ToRuntimeError(errorType int32, msg string) error {
	return NewRuntimeError(errorType, t.Line, t.Lexeme, msg)
}
