package util

type Stack[T any] struct {
	values []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{values: make([]T, 0)}
}

func (s *Stack[T]) Push(value T) {
	s.values = append(s.values, value)
}

func (s *Stack[T]) Pop() T {
	val := s.Peek()
	s.values = s.values[:(len(s.values) - 1)]
	return val
}

func (s *Stack[T]) Peek() T {
	return s.values[len(s.values)-1]
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.values) == 0
}

func (s *Stack[T]) Length() int {
	return len(s.values)
}

func (s *Stack[T]) Get(idx int) T {
	return s.values[idx]
}

func (s *Stack[T]) ForEach(f func(i int, val T) bool) {
	for i := s.Length() - 1; i >= 0; i = i - 1 {
		cont := f(i, s.values[i])
		if !cont {
			return
		}
	}
}
