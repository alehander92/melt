package compiler

import "gitlab.com/alehander42/melt/types"

// ToString helper
// ToString("x") -> &String{Value: "x"}
func ToString(value string) *String {
	return &String{Value: value}
}

// String node
type String struct {
	Value string

	Info
}

// Template node: generate fmt.Sprintf(text, args)
type Template struct {
	Text []string
	Args []Ast

	Info
}

func (self *String) TypeCheck(ctx *Context) error {
	m, err := ctx.Get("string")
	if err != nil {
		m = types.Basic{Label: "string"}
		ctx.Root.Set("string", m)
	}
	self.meltType = m
	return nil
}

func (self *Template) TypeCheck(ctx *Context) error {
	m, err := ctx.Get("string")
	if err != nil {
		m = types.Basic{Label: "string"}
		ctx.Root.Set("string", m)
	}
	self.meltType = m
	for _, w := range self.Args {
		err = w.TypeCheck(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
