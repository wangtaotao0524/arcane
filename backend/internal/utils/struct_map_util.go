package utils

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](vals ...T) Set[T] {
	s := make(Set[T])
	for _, v := range vals {
		s[v] = struct{}{}
	}
	return s
}

func (s Set[T]) Add(vals ...T) {
	for _, v := range vals {
		s[v] = struct{}{}
	}
}

func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) Remove(vals ...T) {
	for _, v := range vals {
		delete(s, v)
	}
}

func (s Set[T]) Items() []T {
	out := make([]T, 0, len(s))
	for v := range s {
		out = append(out, v)
	}
	return out
}

// NewEmptyStructMap returns map[K]struct{}{}
func NewEmptyStructMap[K comparable]() map[K]struct{} {
	return make(map[K]struct{})
}
