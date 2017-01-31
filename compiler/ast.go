package compiler

import "fmt"
import (
	// "errors"
	// "reflect"

	"gitlab.com/alehander42/melt/types"
)

type LocationInfo struct {
	Line   int
	Column int
}

type MType struct {
	ZType types.Type
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
	MType
}

func (self Info) MeltType() types.Type {
	return self.ZType
}

func (self *Info) ChangeMeltType(t types.Type) {
	fmt.Printf("%T %s \n", self, t.ToString())
	self.ZType = t
}

func (self Info) ToString(depth int) string {
	return "self"
}

func (self Info) Location() LocationInfo {
	return self.LocationInfo
}
