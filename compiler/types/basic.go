package types

type Basic struct {
	Label string
}

func (self Basic) ToString() string {
	return self.Label
}

func (self Basic) Accepts(t Type) bool {
	switch other := t.(type) {
	case Basic:
		return self.Label == other.Label
	default:
		return false
	}
}
