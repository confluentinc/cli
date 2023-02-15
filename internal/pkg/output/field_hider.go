package output

import (
	"fmt"
	"reflect"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

type FieldHider struct {
	format Format
	filter *[]string
}

func (h FieldHider) MakeTag(t reflect.Type, i int) reflect.StructTag {
	if *h.filter == nil || utils.Contains(*h.filter, t.Field(i).Name) {
		return t.Field(i).Tag
	}
	return reflect.StructTag(fmt.Sprintf(`%s:"-"`, h.format))
}
