package compiler

func Instantiate(m *Module, ctx *Context) error {
	for _, f := range m.Functions {
		g, ok := ctx.Dependencies[f.Label.Label]
		if ok {
			ExpandDependencies(g, f, m.Functions, ctx)
		}
	}

	return nil
}
