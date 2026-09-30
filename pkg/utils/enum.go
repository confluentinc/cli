package utils

import (
	"fmt"
)

type EnumUtils map[string]any

func (enumUtils EnumUtils) Init(enums ...any) EnumUtils {
	for _, enum := range enums {
		enumUtils[fmt.Sprint(enum)] = enum
	}
	return enumUtils
}
