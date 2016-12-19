package compiler

import (
	"errors"

	"gitlab.com/alehander42/melt/types"
)

//Operator int
type Operator int

const (
	//EqualOp =
	EqualOp Operator = 1
	// NotEqualOp !=
	NotEqualOp Operator = 2
)

//BinaryOperator binary
type BinaryOperator int

const (
	//AddOp +
	AddOp BinaryOperator = 1
	//SubOp -
	SubOp BinaryOperator = 2
	//MultOp *
	MultOp BinaryOperator = 3
	//DivideOp *
	DivideOp BinaryOperator = 4
)

// BinaryOperation node
type BinaryOperation struct {
	Op    BinaryOperator
	Left  *Ast
	Right *Ast

	Info
}

// all are defined for int and float
// + is defined also for strings
// for int/float, return float if there is a float, otherwise int
func (self *BinaryOperation) TypeCheck(ctx *Context) error {
	err := (*self.Right).TypeCheck(ctx)
	if err != nil {
		return err
	}

	err = (*self.Left).TypeCheck(ctx)
	if err != nil {
		return err
	}

	right, ok := (*self.Right).MeltType().(types.Basic)
	left, ok2 := (*self.Left).MeltType().(types.Basic)
	if !ok || !ok2 {
		return errors.New("Expected ints, floats or strings")
	} else if right.Label == "string" {
		if left.Label != "string" || self.Op != AddOp {
			return errors.New("Only string + supported")
		} else {
			self.meltType = right
		}
	} else if right.Label == "int" && left.Label == "int" {
		self.meltType = right
	} else if right.Label == "float" && left.Label == "int" ||
		right.Label == "float" && left.Label == "float" {
		self.meltType = right
	} else if right.Label == "int" && left.Label == "float" {
		self.meltType = left
	} else {
		return errors.New("Types not supported binary")
	}
	return nil
}
