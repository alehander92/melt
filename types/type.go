package types

type Type interface {
	Accepts(Type) bool
	ToString() string
}
