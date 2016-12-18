package compiler

import (
	"strings"
)

func Indent(depth int) string {
	return strings.Repeat("  ", depth)
}
