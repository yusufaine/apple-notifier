package set

// This is not thread-safe
type Stringset struct {
	set map[string]struct{}
}

func NewStringset() *Stringset {
	return &Stringset{
		set: make(map[string]struct{}),
	}
}

func FromStrings(els ...string) *Stringset {
	s := NewStringset()
	for _, el := range els {
		s.set[el] = struct{}{}
	}

	return s
}

func (s *Stringset) Add(els ...string) {
	for _, el := range els {
		s.set[el] = struct{}{}
	}
}

func (s *Stringset) Contains(el string) bool {
	_, ok := s.set[el]
	return ok
}

func (s *Stringset) Len() int {
	return len(s.set)
}

func (s *Stringset) Remove(els ...string) {
	for _, el := range els {
		delete(s.set, el)
	}
}

func (s *Stringset) Slice() []string {
	keys := make([]string, 0, len(s.set))
	for k := range s.set {
		keys = append(keys, k)
	}
	return keys
}
