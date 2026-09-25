// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package errors

import (
	"fmt"

	"github.com/fatih/color"
)

// FlagError describes a problem with the value of a CLI flag.
type FlagError struct {
	Type ErrorType
	Msg  string
}

// Error implements the error interface.
func (e *FlagError) Error() string {
	switch e.Type {
	case TypeHint:
		return fmt.Sprintf("hint: %s", e.Msg)
	case TypeError:
		return fmt.Sprintf("error: %s", e.Msg)
	default:
		return e.Msg
	}
}

// HandlerString returns a colored error string for CLI display.
func (e *FlagError) HandlerString() string {
	switch e.Type {
	case TypeHint:
		return color.YellowString(e.Error())
	case TypeError:
		return color.RedString(e.Error())
	default:
		return e.Error()
	}
}

// NewFlagFileError creates an error for a flag whose file value cannot be read.
func NewFlagFileError(flag, path string, err error) error {
	msg := fmt.Sprintf("cannot read --%s file %q: %s", flag, path, err.Error())
	return &FlagError{
		Type: TypeError,
		Msg:  msg,
	}
}
