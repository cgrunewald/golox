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
	E_DIVIDE_BY_ZERO
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
