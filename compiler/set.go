package compiler

import (
	"errors"

	"gitlab.com/alehander42/melt/types"
)

type Set struct {
	Label *Label
	Value *Ast

	Info
}

func (s *Set) TypeCheck(ctx *Context) error {
	err := (*s.Value).TypeCheck(ctx)
	if err != nil {
		return err
	}

	target, err := ctx.Get(s.Label.Label)
	if err != nil {
		ctx.Set(s.Label.Label, (*s.Value).MeltType())
		s.Label.ZType = (*s.Value).MeltType()
		s.ZType = types.Empty{}
		return nil
	} else {
		if target.Accepts((*s.Value).MeltType()) {
			s.Label.ZType = target
			s.ZType = types.Empty{}
			return nil
		} else {
			return errors.New("fail")
		}
	}
}
