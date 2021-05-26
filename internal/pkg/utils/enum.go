package utils

import (
	"fmt"
)

type EnumUtils map[string]interface{}

func (enumUtils EnumUtils) Init(enums ...interface{}) EnumUtils {
	for _, enum := range enums {
		enumUtils[fmt.Sprintf("%v", enum)] = enum
	}
	return enumUtils
}
