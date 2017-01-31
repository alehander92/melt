package compiler

import "errors"

// Cmp node
type Cmp struct {
	Op    Operator
	Left  Ast
	Right Ast

	Info
}

func (self *Cmp) TypeCheck(ctx *Context) error {
	err := self.Right.TypeCheck(ctx)
	if err != nil {
		return err
	}
	err = self.Left.TypeCheck(ctx)
	if err != nil {
		return err
	}

	if self.Left.MeltType().Accepts(self.Right.MeltType()) {
		m, _ := ctx.Get("bool")
		self.ZType = m
		return nil
	} else {
		return errors.New("Left doesn't match right")
	}
}
