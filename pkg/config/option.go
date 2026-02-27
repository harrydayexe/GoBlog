package config

type Option[T any] interface {
	AsOption() T
}
