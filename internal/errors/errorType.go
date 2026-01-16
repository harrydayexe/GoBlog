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
