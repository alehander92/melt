package compiler

import (
	"errors"
	"fmt"

	"gitlab.com/alehander42/melt/compiler/types"
)

type ForIn struct {
	Index    []Label
	Sequence *Ast
	Code     *Code

	Info
}

func (f *ForIn) TypeCheck(ctx *Context) error {
	err := (*f.Sequence).TypeCheck(ctx)
	if err != nil {
		return err
	}

	duck, ok := (*f.Sequence).MeltType().(types.Duck)
	if !ok {
		return errors.New("Doesn't support for")
	}

	var index Label
	var value Label

	if len(f.Index) == 2 {
		index = f.Index[0]
		value = f.Index[1]
	} else {
		value = f.Index[0]
	}

	if len(f.Index) == 2 {
		_, err = ctx.Get(index.Label)
		if err == nil {
			return fmt.Errorf("Can't redefine index %s", index.Label)
		}
	}
	_, err = ctx.Get(value.Label)
	if err == nil {
		return fmt.Errorf("Can't redefine index %s", value.Label)
	}

	codeCtx := NewContextIn(ctx)
	if s, ok := duck.(types.MapBuiltin); ok {
		codeCtx.Set(index.Label, s.Key)
		codeCtx.Set(value.Label, s.Value)
	} else {
		if len(f.Index) == 2 {
			number, _ := ctx.Get("int")
			codeCtx.Set(index.Label, number)
		}
		if t, ok := duck.(types.SliceBuiltin); ok {
			codeCtx.Set(value.Label, t.Element)
		} else {
			u, ok := types.Accepts(duck, "Begin")
			v, ok2 := types.Accepts(duck, "Next")
			if !ok || !ok2 {
				return errors.New("Needs to define Begin() *T and Next() *T")
			}

			if u2, ok := IterableMethod(u.Function); !ok {
				return errors.New("Invalid Begin()")
			} else if v2, ok := IterableMethod(v.Function); !ok {
				return errors.New("Invalid Next()")
			} else if !u2.Object.Accepts(v2.Object) {
				return errors.New("Invalid Next() type")
			} else {
				codeCtx.Set(value.Label, v2.Object)
			}
		}
	}
	err = f.Code.TypeCheck(codeCtx)
	if err != nil {
		return err
	}
	return nil
}

func IterableMethod(u types.Function) (types.Pointer, bool) {
	u2, ok := u.Return.(types.Pointer)
	ok = ok && len(u.Args) == 0 && u.Error == types.Correct
	if ok {
		return u2, ok
	} else {
		return types.Pointer{}, false
	}
}

type ForLoop struct {
	Index *Label
	Begin *Ast
	End   *Ast
	Code  *Code

	Info
}

func (self *ForLoop) TypeCheck(ctx *Context) error {
	err := (*self.Begin).TypeCheck(ctx)
	if err != nil {
		return err
	}

	err = (*self.End).TypeCheck(ctx)
	if err != nil {
		return err
	}

	_, err = ctx.Get(self.Index.Label)
	if err != nil {
		return errors.New("For index already defined")
	}

	if begin, ok := (*self.Begin).MeltType().(types.Basic); ok {
		if begin.Label != "int" {
			return errors.New("For begin should be an int")
		}

		if end, ok := (*self.End).MeltType().(types.Basic); ok {
			if end.Label != "int" {
				return errors.New("For end should be an int")
			}
		} else {
			return errors.New("For end should be an int")
		}

		forCtx := NewContextIn(ctx)
		forCtx.Set(self.Index.Label, begin)
		self.Index.meltType = begin
		err = self.Code.TypeCheck(forCtx)
		if err != nil {
			return err
		}
	} else {
		return errors.New("For begin should be an int")
	}
	return nil
}
