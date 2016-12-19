package types

import "fmt"

type Error struct {
	Label string
}

func (self Error) ToString() string {
	return fmt.Sprintf("Error.%s", self.Label)
}

func (self Error) Accepts(t Type) bool {
	return false
}
