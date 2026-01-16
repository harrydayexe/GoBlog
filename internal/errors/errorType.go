package errors

type ErrorType int

const (
	TypeHint ErrorType = iota
	TypeError
)

func (et ErrorType) IsFatalError() bool {
	return et == TypeError
}
