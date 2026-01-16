package errors

import (
	"fmt"

	"github.com/fatih/color"
)

type InputDirectoryError struct {
	Type ErrorType
	Msg  string
}

func (e *InputDirectoryError) Error() string {
	switch e.Type {
	case TypeHint:
		return fmt.Sprintf("hint: %s", e.Msg)
	case TypeError:
		return fmt.Sprintf("error: %s", e.Msg)
	default:
		return fmt.Sprintf("%s", e.Msg)
	}
}

func (e *InputDirectoryError) HandlerString() string {
	switch e.Type {
	case TypeHint:
		return color.YellowString(e.Error())
	case TypeError:
		return color.RedString(e.Error())
	default:
		return e.Error()
	}
}

func NewDirectoryInaccessibleError(err error) error {
	msg := fmt.Sprintf("cannot access directory: %s", err.Error())
	return &InputDirectoryError{
		Type: TypeError,
		Msg:  msg,
	}
}

func NewNotADirectoryError(path string) error {
	msg := fmt.Sprintf("path is not a directory: %s", path)
	return &InputDirectoryError{
		Type: TypeError,
		Msg:  msg,
	}
}

func NewPathNotSpecifiedError() error {
	msg := "please specify a path"
	return &InputDirectoryError{
		Type: TypeHint,
		Msg:  msg,
	}
}
