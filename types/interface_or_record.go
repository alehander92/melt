package types

type InterfaceOrRecord interface {
	ToString() string
	Accepts(Type) bool
	AcceptsFunction(Function) bool
}
