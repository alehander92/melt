package compiler

import "fmt"

// Code node
// Has only one member with a list of the nodes
type Code struct {
	E []Ast

	Info
}

func (self *Code) TypeCheck(ctx *Context) error {
	for _, expression := range self.E {
		err := expression.TypeCheck(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Code) ToString(depth int) string {
	return fmt.Sprintf("%sCode", Indent(depth))
}
