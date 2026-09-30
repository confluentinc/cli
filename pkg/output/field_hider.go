package output

import (
	"fmt"
	"reflect"
	"slices"
)

type FieldHider struct {
	format Format
	filter *[]string
}

func (h FieldHider) MakeTag(t reflect.Type, i int) reflect.StructTag {
	if *h.filter == nil || slices.Contains(*h.filter, t.Field(i).Name) {
		return t.Field(i).Tag
	}
	return reflect.StructTag(fmt.Sprintf(`%s:"-"`, h.format))
}
