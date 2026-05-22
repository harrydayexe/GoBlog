// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package errors

import (
	"fmt"

	"github.com/fatih/color"
)

type InputDirectoryError struct {
	Type ErrorType
	Msg  string
}

// Error implements the error interface.
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

// HandlerString returns a colored error string for CLI display.
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

// NewDirectoryInaccessibleError creates an error for when a directory cannot be accessed.
func NewDirectoryInaccessibleError(err error) error {
	msg := fmt.Sprintf("cannot access directory: %s", err.Error())
	return &InputDirectoryError{
		Type: TypeError,
		Msg:  msg,
	}
}

// NewNotADirectoryError creates an error for when a path is not a directory.
func NewNotADirectoryError(path string) error {
	msg := fmt.Sprintf("path is not a directory: %s", path)
	return &InputDirectoryError{
		Type: TypeError,
		Msg:  msg,
	}
}

// NewPathNotSpecifiedError creates an error for when a path is not specified.
func NewPathNotSpecifiedError() error {
	msg := "please specify a path"
	return &InputDirectoryError{
		Type: TypeHint,
		Msg:  msg,
	}
}
