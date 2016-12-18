package compiler

// Import node
type Import struct {
	Package string
	Alias   string

	Info
}

// MeltImport node
type MeltImport struct {
	Go   []Import
	Melt []Import

	Info
}

func (m *MeltImport) TypeCheck(ctx *Context) error {
	return nil
}

func (i *Import) TypeCheck(ctx *Context) error {
	return nil
}
