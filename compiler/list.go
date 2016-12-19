package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/types"
)

type List struct {
	Elements []Ast

	Info
}

func (l *List) TypeCheck(ctx *Context) error {
	var item types.Type
	if len(l.Elements) == 0 {
		return errors.New("[] needs as")
	}
	for i, element := range l.Elements {
		err := element.TypeCheck(ctx)
		if err != nil {
			return err
		}
		if i == 0 {
			item = element.MeltType()
		} else if !item.Accepts(element.MeltType()) {
			return fmt.Errorf("List expects %s", item.ToString())
		}
	}
	l.meltType = types.SliceBuiltin{Element: item}
	return nil
}
