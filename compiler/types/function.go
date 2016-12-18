package types

import (
	"fmt"
	"strings"
)

//ErrorFunction enum
type ErrorFunction int

const (
	Fail    ErrorFunction = 1
	Maybe   ErrorFunction = 2
	Correct ErrorFunction = 3
)

func Alexander(e ErrorFunction) string {
	if e == Fail {
		return "!"
	} else if e == Maybe {
		return "?"
	} else {
		return ""
	}
}

type Function struct {
	Args         []Type
	Return       Type
	GenericVars  []GenericVar
	InstanceVars []Type
	Error        ErrorFunction
}

type Method struct {
	Label    string
	Function Function
}

func (self Function) ToString() string {
	var args []string
	for _, arg := range self.Args {
		args = append(args, arg.ToString())
	}
	argsString := strings.Join(args, ",")
	return fmt.Sprintf("(%s -> %s)%s", argsString, self.Return.ToString(), Alexander(self.Error))
}

func (self Function) Accepts(t Type) bool {
	switch other := t.(type) {
	case Interface:
		return other.Label == "Callable"
	case Function:
		if !self.Return.Accepts(other.Return) {
			return false
		}
		if self.Error == Correct && other.Error != self.Error ||
			self.Error == Fail && other.Error != self.Error {
			return false
		}
		if len(self.GenericVars) != len(other.GenericVars) {
			return false
		}
		if len(self.Args) != len(other.Args) {
			return false
		}
		for i := range self.Args {
			if self.Args[i].Accepts(other.Args[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (self Function) IsGeneric() bool {
	return len(self.GenericVars) > 0
}

func (self Function) Vars() []GenericVar {
	return self.GenericVars
}

func (self Function) IVars() []Type {
	return self.InstanceVars
}
