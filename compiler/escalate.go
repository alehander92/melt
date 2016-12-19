package compiler

import (
	"fmt"
	"strings"

	"gitlab.com/alehander42/melt/types"
)

// Escalate node
type Escalate struct {
	Args []*Label

	Info
}

func (self *Escalate) ToString(depth int) string {
	s := []string{}
	for _, arg := range self.Args {
		s = append(s, arg.ToString(depth+1))
	}
	return fmt.Sprintf("%sEscalate:\n%s", Indent(depth), strings.Join(s, "\n"))
}

func (self *Escalate) TypeCheck(ctx *Context) error {
	for _, arg := range self.Args {
		err := self.typeCheckSingle(arg, ctx)
		if err != nil {
			return err
		}
	}

	self.meltType = types.Nil{}
	return nil
}

func (self *Escalate) typeCheckSingle(arg *Label, ctx *Context) error {
	label := arg.Label
	if arg.Label[len(arg.Label)-1] == '!' {
		label = arg.Label[:len(arg.Label)-1]
	} else if arg.Label[len(arg.Label)-1] == '?' {
		label = arg.Label[:len(arg.Label)-1]
	} else {
		label = arg.Label
	}

	m, err := ctx.Get(label)
	if err != nil {
		return fmt.Errorf("escalate %s is not defined", label)
	} else {
		if f, ok := m.(types.Function); ok {
			if f.Error == types.Correct {
				return fmt.Errorf("%s can't return an error", label)
			} else if f.Error == types.Maybe {
				if ctx.Z == types.Correct {
					ctx.Z = types.Maybe
				}
			} else {
				ctx.Z = types.Fail
			}
		} else {
			return fmt.Errorf("%s is not a function", label)
		}
	}
	delete(*ctx.Unhandled, label)
	return nil
}
