package domain

// Option represents an explicit nullable domain value without importing
// persistence null types or using public pointers.
type Option[T any] struct {
	value T
	set   bool
}

func Some[T any](value T) Option[T] {
	return Option[T]{value: value, set: true}
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func (o Option[T]) Get() (T, bool) {
	return o.value, o.set
}

func (o Option[T]) IsSome() bool {
	return o.set
}

func (o Option[T]) IsNone() bool {
	return !o.set
}
