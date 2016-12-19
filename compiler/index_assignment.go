package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

// Index assignment
type IndexAssignment struct {
	Collection *Ast
	Index      *Ast
	Value      *Ast

	Info
}

func (self *IndexAssignment) ToString(depth int) string {
	return fmt.Sprintf("%sCollection:\n%s\n%s\n%s",
		Indent(depth),
		(*self.Collection).ToString(depth+1),
		(*self.Index).ToString(depth+1),
		(*self.Value).ToString(depth+1))
}

func (self *IndexAssignment) TypeCheck(ctx *Context) error {
	err := (*self.Collection).TypeCheck(ctx)
	if err != nil {
		return err
	}
	err = (*self.Index).TypeCheck(ctx)
	if err != nil {
		return err
	}
	err = (*self.Value).TypeCheck(ctx)
	if err != nil {
		return err
	}
	self.meltType = (*self.Collection).MeltType()

	switch object := (*self.Collection).MeltType().(type) {
	case types.SliceBuiltin:
		if j, qk := (*self.Index).MeltType().(types.Basic); qk {
			if j.Label == "int" {
				if object.Element.Accepts((*self.Value).MeltType()) {
					return nil
				} else {
					return fmt.Errorf("%s doesn't accept %s",
						object.ToString(),
						(*self.Value).MeltType().ToString())
				}
			} else {
				return errors.New("Slice expect int")
			}
		} else {
			return errors.New("Slice expect basic")
		}
	case types.MapBuiltin:
		if object.Key.Accepts((*self.Index).MeltType()) && object.Value.Accepts((*self.Value).MeltType()) {
			return nil
		} else {
			return errors.New("Map confused")
		}
	default:
		return errors.New("Index only supported for slices and maps")
	}
}
