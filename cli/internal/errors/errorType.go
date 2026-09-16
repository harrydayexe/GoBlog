// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package errors

type ErrorType int

const (
	TypeHint ErrorType = iota
	TypeError
)

// IsFatalError returns true if the error type is fatal and should cause the program to exit.
func (et ErrorType) IsFatalError() bool {
	return et == TypeError
}
