package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

type Make struct {
	Type types.Type
	Args []Ast

	Info
}

func (m *Make) TypeCheck(ctx *Context) error {
	// fmt.Printf("%T\n", m.Args[0])
	slice, ok := m.Type.(types.SliceBuiltin)
	if ok {
		if len(m.Args) < 1 {
			return errors.New("make expect 1 arg")
		} else {
			arg := m.Args[0]
			err := arg.TypeCheck(ctx)
			if err != nil {
				return err
			}

			argType, ok := arg.MeltType().(types.Basic)
			if ok && argType.Label == "int" {
				m.meltType = slice
			} else {
				return errors.New("please pass an int")
			}
		}
	} else {
		n, ok := m.Type.(types.MapBuiltin)
		if ok {
			m.meltType = n
		} else {
			return fmt.Errorf("make slice map %s", m.Type.ToString())
		}
	}
	return nil
}
