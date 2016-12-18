package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/compiler/types"
)

// Return node
type Return struct {
	Value *Ast

	Info
}

func (r *Return) TypeCheck(ctx *Context) error {
	err := (*r.Value).TypeCheck(ctx)
	if err != nil {
		return err
	}

	if !ctx.ReturnType.Accepts((*r.Value).MeltType()) {
		return fmt.Errorf("Return type %s != %s", ctx.ReturnType.ToString(), (*r.Value).MeltType().ToString())
	} else {
		r.meltType = types.Empty{}
		return nil
	}
}

// ReturnError node
type ReturnError struct {
	Value *Ast

	Info
}

func (r *ReturnError) TypeCheck(ctx *Context) error {
	if ctx.Z == types.Correct {
		return errors.New("Function has to be marked with ? or ! to fail")
	}
	// if Maybe handler?

	err := (*r.Value).TypeCheck(ctx)
	if err != nil {
		return err
	}

	s, ok := (*r.Value).MeltType().(types.Basic)
	if ok {
		if s.Label != "string" {
			return errors.New("!! expects string")
		} else {
			r.meltType = types.Empty{}
			return nil
		}
	} else {
		return errors.New("!! expects simple")
	}
}

// Error node
type Error struct {
	Label *Label

	Info
}

func (e *Error) TypeCheck(ctx *Context) error {
	if e.Label.Label != "err" {
		return errors.New("Only $err defined")
	}

	e.meltType = types.Error{}
	return nil
}
