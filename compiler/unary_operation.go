package compiler

import (
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

//UnaryOperator binary
type UnaryOperator int

const (
	//ZeroOp +
	PlusOp UnaryOperator = 1
	//UnaryOp -
	MinusOp UnaryOperator = 0
)

// UnaryOperation -
type UnaryOperation struct {
	Op         UnaryOperator
	Expression *Ast

	Info
}

func (self *UnaryOperation) TypeCheck(ctx *Context) error {
	err := (*self.Expression).TypeCheck(ctx)
	if err != nil {
		return err
	}

	if m, ok := (*self.Expression).MeltType().(types.Basic); ok {
		if m.Label == "float" || m.Label == "int" {
			self.ZType = m
			return nil
		}
	}
	return fmt.Errorf("%s expected float or int, got: %s",
		self.OpText(),
		(*self.Expression).MeltType().ToString())
}

func (self *UnaryOperation) OpText() string {
	if self.Op == PlusOp {
		return "+"
	} else {
		return "-"
	}
}
