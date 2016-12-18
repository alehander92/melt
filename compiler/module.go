package compiler

import (
	"fmt"
)

// Module node
// A single file corresponds to it
type Module struct {
	Package    string
	Imports    MeltImport
	Functions  []Function
	Interfaces []Interface
	Records    []Record

	Info
}

func (self *Module) TypeCheck(ctx *Context) error {
	fmt.Println("%s\n", self)
	err := ctx.CollectTypes(*self)
	if err != nil {
		return err
	}

	for _, f := range self.Functions {
		err = f.TypeCheck(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
