package compiler

// Type node
type Type struct {
	Type string
}

// ToType helper
func ToType(type_ string) Type {
	return Type{Type: type_}
}

func (*Type) TypeCheck(*Context) error {
	return nil
}

type Empty struct {
}
