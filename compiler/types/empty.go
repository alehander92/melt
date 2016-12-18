package types

type Empty struct{}

func (self Empty) ToString() string {
	return "empty"
}

// Accepts can't really happen for empty
func (self Empty) Accepts(t Type) bool {
	return false
}
