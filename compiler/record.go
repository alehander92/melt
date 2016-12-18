package compiler

type Record struct {
	Label  *Label
	Fields []Field

	Info
}

type Field struct {
	Label *Label

	Info
}

func (r *Record) TypeCheck(ctx *Context) error {
	return nil
}

func (f *Field) TypeCheck(ctx *Context) error {
	return nil
}
