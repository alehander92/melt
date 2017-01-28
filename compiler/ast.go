package compiler

import (
	// "errors"
	// "fmt"
	// "reflect"

	"gitlab.com/alehander42/melt/types"
)

type LocationInfo struct {
	Line   int
	Column int
}

type Ast interface {
	Location() LocationInfo
	ToString(int) string
	MeltType() types.Type
	ChangeMeltType(types.Type)
	TypeCheck(ctx *Context) error
}

type Info struct {
	LocationInfo
	meltType types.Type
}

func (self Info) MeltType() types.Type {
	return self.meltType
}

func (self *Info) ChangeMeltType(t types.Type) {
	self.meltType = t
}

func (self Info) ToString(depth int) string {
	return "self"
}

func (self Info) Location() LocationInfo {
	return self.LocationInfo
}
