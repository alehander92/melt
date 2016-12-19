package main

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"os"

	"gitlab.com/alehander42/melt/compiler"
	"gitlab.com/alehander42/melt/generator"
)

func main() {
	if len(os.Args) < 2 {
		problem("Please: filename")
	}
	source, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		problem(fmt.Sprintf("File: %s", err))
	}

	ast, err := compiler.Parse(string(source))
	if err != nil {
		problem(fmt.Sprintf("Parser: %s", err))
	}

	ctx := compiler.NewContext()
	ctx.LoadBuiltinTypes()

	err = ast.TypeCheck(&ctx)
	if err != nil {
		problem(fmt.Sprintf("%s", err))
	}

	fileSet, file, err := generator.Generate(ast, &ctx)
	if err != nil {
		problem(fmt.Sprintf("%s", err))
	}

	e, err := os.Create(fmt.Sprintf("%s.go", os.Args[1]))
	if err != nil {
		problem("Can't write")
	}
	defer e.Close()

	err = format.Node(e, fileSet, file)
	// err = ioutil.WriteFile(, []byte(text), 0644)
	if err != nil {
		problem(fmt.Sprintf("%s", err))
	}
}

func problem(message string) {
	fmt.Printf("ERROR:\n  %s\n", message)
	os.Exit(1)
}
