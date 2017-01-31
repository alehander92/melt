package compiler

import "gitlab.com/alehander42/melt/types"

type Nil struct {
	Value string

	Info
}

func (n *Nil) TypeCheck(ctx *Context) error {
	n.ZType = types.Empty{}
	return nil
}
