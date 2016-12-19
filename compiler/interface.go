package compiler

import "gitlab.com/alehander42/melt/types"

// Interface node
type Interface struct {
	Label   *Label
	Methods []InterfaceMethod

	Info
}

// InterfaceMethod node
type InterfaceMethod struct {
	Label *Label
	Type  types.Function

	Info
}

func (i *Interface) TypeCheck(ctx *Context) error {
	return nil
}

func (i *InterfaceMethod) TypeCheck(ctx *Context) error {
	return nil
}
