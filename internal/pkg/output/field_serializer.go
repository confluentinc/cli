package output

import (
	"fmt"
	"reflect"
)

type FieldSerializer struct {
	format Format
}

func (s FieldSerializer) MakeTag(t reflect.Type, i int) reflect.StructTag {
	if val, ok := t.Field(i).Tag.Lookup("serialized"); ok {
		return reflect.StructTag(fmt.Sprintf(`%s:"%s"`, s.format.String(), val))
	}
	return t.Field(i).Tag
}
