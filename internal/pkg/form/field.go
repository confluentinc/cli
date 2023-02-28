package form

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type Field struct {
	ID           string
	Prompt       string
	DefaultValue any
	IsYesOrNo    bool
	IsHidden     bool
	Regex        string
	RequireYes   bool
}

func (f Field) read(prompt Prompt) (string, error) {
	var val string
	var err error

	if f.IsHidden {
		val, err = prompt.ReadLineMasked()
	} else {
		val, err = prompt.ReadLine()
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (f Field) validate(val string) (any, error) {
	if f.IsYesOrNo {
		switch strings.ToUpper(val) {
		case "Y", "YES":
			return true, nil
		case "N", "NO":
			return false, nil
		}
		return false, fmt.Errorf(errors.InvalidChoiceMsg, val)
	}

	if val == "" && f.DefaultValue != nil {
		return f.DefaultValue, nil
	}

	if f.Regex != "" {
		re, _ := regexp.Compile(f.Regex)
		if match := re.MatchString(val); !match {
			return nil, fmt.Errorf(errors.InvalidInputFormatErrorMsg, val, f.ID)
		}
	}

	return val, nil
}

func (f Field) String() string {
	out := f.Prompt
	if f.IsYesOrNo {
		out += " (y/n)"
	}
	out += ": "

	if f.DefaultValue != nil {
		out += fmt.Sprintf("(%v) ", f.DefaultValue)
	}

	return out
}
