package generator

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"

	comp "gitlab.com/alehander42/melt/compiler"
)

// Generate returns go ast which can be then compiled to code
func Generate(meltAst comp.Module, ctx *comp.Context) (*token.FileSet, *ast.File, error) {
	// b()
	f := token.NewFileSet()
	a, err := GenerateModule(meltAst, ctx)
	if err != nil {
		return nil, nil, err
	}

	return f, a, nil
}

func b() {
	a := `
package main

func MapIntIntSequenceSlice(handler (func(int) int), sequence []int) []int {
	result := make([]int, sequence.Length())
	for i, item := range sequence {
		result[i] = handler(item)
	}

	return result
}

func Double(number int) int {
	return number * 2
}

type A struct {
	x int
}


func main() {
	fmt.Printf("%s\n", MapIntIntSequenceSlice(Double, []int{2}))
}
`
	b := token.NewFileSet() // positions are relative to fset
	c, err := parser.ParseFile(b, "", a, 0)
	if err != nil {
		panic(err)
	}

	ast.Print(b, c)
	s := format.Node(os.Stdout, b, c)
	// Print the AST.
	fmt.Printf("%s\n", s)
}
