package types

import "fmt"

type MapBuiltin struct {
	Value   Type
	Key     Type
	methods []Method
}

func (m MapBuiltin) IsGeneric() bool {
	v, ok := m.Value.(Generic)
	if ok && v.IsGeneric() {
		return true
	} else {
		k, ok := m.Key.(Generic)
		if ok && k.IsGeneric() {
			return true
		}
	}
	return false
}

func (m MapBuiltin) Methods() []Method {
	return m.methods
}

func (m MapBuiltin) Vars() []GenericVar {
	v, ok := m.Value.(Generic)
	vars := []GenericVar{}
	values := []GenericVar{}
	if ok {
		values = v.Vars()
		vars = append(vars, values...)
	}
	k, ok := m.Key.(Generic)
	if ok {
		variables := k.Vars()
		for _, w := range variables {
			fix := false
			for _, v := range values {
				if w.Label == v.Label {
					fix = true
					break
				}
			}

			if !fix {
				vars = append(vars, w)
			}
		}
	}
	return vars
}

func (m MapBuiltin) IVars() []Type {
	return []Type{}
}

func (m MapBuiltin) ToString() string {
	return fmt.Sprintf("map[%s]%s", m.Key.ToString(), m.Value.ToString())
}

func (m MapBuiltin) Accepts(t Type) bool {
	if a, ok := t.(MapBuiltin); ok {
		return m.Key.Accepts(a.Key) && m.Value.Accepts(a.Value)
	} else {
		return false
	}
}
