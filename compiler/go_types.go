package compiler

import (
	"errors"
	"fmt"
	go_types "go/types"

	"gitlab.com/alehander42/melt/compiler/types"
)

func TranslateType(goType go_types.Type) (types.Type, error) {
	switch g := goType.(type) {
	case *go_types.Basic:
		return types.Basic{Label: g.Name()}, nil
	default:
		return types.Basic{}, errors.New(fmt.Sprintf("No %s", "type"))
	}
}
