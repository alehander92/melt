package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/compiler/types"
)

// If node
type If struct {
	Test      *Cmp
	Code      *Code
	Otherwise *Code
}

func (self *If) TypeCheck(ctx *Context) error {
	t := self.Test.TypeCheck(ctx)
	if t != nil {
		return t
	} else {
		if a, ok := self.Test.MeltType().(types.Basic); ok {
			if a.Label != "bool" {
				return errors.New("if expect bool test")
			}
		} else {
			return errors.New("if expect bool test")
		}
	}

	err := self.Code.TypeCheck(ctx)
	if err != nil {
		return err
	}
	other := self.Otherwise.TypeCheck(ctx)
	if other != nil {
		return other
	}
	return nil
}

func (self *If) ToString(depth int) string {
	return fmt.Sprintf("%sIf:%s",
		Indent(depth),
		self.Test.ToString(depth+1))
}
