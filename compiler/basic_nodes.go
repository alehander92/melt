package compiler

import (
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

//Bool node
type Bool struct {
	Value bool

	Info
}

// Integer node
type Integer struct {
	Value int64

	Info
}

// Float node
type Float struct {
	Value float64

	Info
}

func (self *Bool) TypeCheck(ctx *Context) error {
	m, err := ctx.Get("bool")
	if err != nil {
		m = types.Basic{Label: "bool"}
		ctx.Root.Set("bool", m)
	}
	self.ZType = m
	return nil
}

func (self *Integer) TypeCheck(ctx *Context) error {
	m, err := ctx.Get("int")
	if err != nil {
		m := types.Basic{Label: "int"}
		ctx.Root.Set("int", m)
	}
	self.ZType = m
	return nil
}

func (self *Float) TypeCheck(ctx *Context) error {
	m, err := ctx.Get("float")
	if err != nil {
		m := types.Basic{Label: "float"}
		ctx.Root.Set("float", m)
	}
	self.ZType = m
	return nil
}

func (self *Bool) ToString(depth int) string {
	return fmt.Sprintf("%sBool:%s", Indent(depth), self.Value)
}

func (self *Integer) ToString(depth int) string {
	return fmt.Sprintf("%sInteger:%d", Indent(depth), self.Value)
}

func (self *Float) ToString(depth int) string {
	return fmt.Sprintf("%sFloat:%f", Indent(depth), self.Value)
}

// ToInteger helper
// ToInteger(s) -> &Integer{Value: s}
func ToInteger(value int64) *Integer {
	return &Integer{Value: value}
}

func ToFloat(value float64) *Float {
	return &Float{Value: value}
}

func ToBool(value bool) *Bool {
	return &Bool{Value: value}
}
