package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/compiler/types"
)

type On struct {
	Label   *Label
	Handler *Code

	Info
}

func (o *On) TypeCheck(ctx *Context) error {
	label, err := ctx.Get(o.Label.Label)
	if err != nil {
		return err
	}

	if function, ok := label.(types.Function); ok {
		if function.Error != types.Correct {
			_, ok := (*ctx.Unhandled)[o.Label.Label]
			if ok {
				return errors.New("Already handled error")
			} else {
				system := NewContextIn(ctx)
				delete(*ctx.Unhandled, o.Label.Label)
				system.Set("$err", types.Error{Label: "err"})
				err = o.Handler.TypeCheck(system)
				if err != nil {
					return err
				}
				return nil
			}
		} else {
			return fmt.Errorf("%s can't return an error", o.Label.Label)
		}
	} else {
		return fmt.Errorf("%s is not a function", o.Label.Label)
	}
}
