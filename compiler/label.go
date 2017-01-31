package compiler

import (
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

// Label node
type Label struct {
	Label string

	Info
}

func (self *Label) ToString(depth int) string {
	return fmt.Sprintf("%sLabel:%s", Indent(depth), self.Label)
}

func (self *Label) TypeCheck(ctx *Context) error {
	fail := self.Label[len(self.Label)-1]
	var label string
	if fail == '!' || fail == '?' {
		label = self.Label[:len(self.Label)-1]
	} else {
		label = self.Label
	}

	m, err := ctx.Get(label)
	if err != nil {
		return fmt.Errorf("%s is not defined on %d", label, self.Location().Line)
	}

	n, ok := m.(types.Function)
	if (fail == '!' || fail == '?') && !ok {
		return fmt.Errorf("%s is not a function", label)
	}
	if ok && fail != '!' && fail != '?' && (n.Error == types.Fail) {
		return fmt.Errorf("%s needs %s", label, types.Alexander(n.Error))
	}

	if ok && fail == '!' && n.Error != types.Fail {
		return fmt.Errorf("%s shouldn't be !, but %s", label, types.Alexander(n.Error))
	}

	if ok && fail == '?' && n.Error != types.Maybe {
		return fmt.Errorf("%s shouldn't be ?, but %s", label, types.Alexander(n.Error))
	}

	self.ZType = m
	return nil
}

// ToLabel helper
func ToLabel(label string) *Label {
	return &Label{Label: label}
}
