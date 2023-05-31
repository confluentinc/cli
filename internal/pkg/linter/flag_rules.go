package linter

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/hashicorp/go-multierror"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagRule func(flag *pflag.Flag, cmd *cobra.Command) error

// RequireFlagRealWords checks that a flag uses delimited-real-words, not --smushcaseflags
func RequireFlagRealWords(delim rune) FlagRule {
	return func(flag *pflag.Flag, cmd *cobra.Command) error {
		var issues *multierror.Error
		for _, w := range strings.Split(flag.Name, string(delim)) {
			if ok := vocab.Spell(w); !ok {
				issue := fmt.Errorf("flag name should consist of delimited real english words for --%s on `%s` - unknown %s", flag.Name, FullCommand(cmd), w)
				issues = multierror.Append(issues, issue)
			}
		}
		return issues.ErrorOrNil()
	}
}

// RequireFlagDelimiter checks that a flag uses a specified delimiter at most maxCount times
func RequireFlagDelimiter(delim rune, maxCount int) FlagRule {
	return func(flag *pflag.Flag, cmd *cobra.Command) error {
		countDelim := 0
		for _, l := range flag.Name {
			if l == delim {
				countDelim++
				if countDelim > maxCount {
					return fmt.Errorf("flag name must only have %d delimiter (\"%c\") for --%s on `%s`", maxCount, delim, flag.Name, FullCommand(cmd))
				}
			}
		}
		return nil
	}
}

// RequireFlagCharacters checks that a flag consists only of letters and a delimiter
func RequireFlagCharacters(delim rune) FlagRule {
	return func(flag *pflag.Flag, cmd *cobra.Command) error {
		for _, l := range flag.Name {
			if !unicode.IsLetter(l) && l != delim {
				return fmt.Errorf("flag name must be letters and delim (\"%c\") for --%s on `%s`", delim, flag.Name, FullCommand(cmd))
			}
		}
		return nil
	}
}

// RequireFlagUsageEndWithPunctuation checks that a flag description ends with a period
func RequireFlagUsageEndWithPunctuation(flag *pflag.Flag, cmd *cobra.Command) error {
	if len(flag.Usage) > 0 && flag.Usage[len(flag.Usage)-1] != '.' {
		return fmt.Errorf("flag usage doesn't end with punctuation for --%s on `%s`", flag.Name, FullCommand(cmd))
	}
	return nil
}

func RequireFlagUsageMessage(flag *pflag.Flag, cmd *cobra.Command) error {
	if len(flag.Usage) == 0 {
		return fmt.Errorf("flag must provide help message for --%s on `%s`", flag.Name, FullCommand(cmd))
	}
	return nil
}

// RequireFlagUsageCapitalized checks that a flag description starts with a capital letter
func RequireFlagUsageCapitalized(properNouns []string) FlagRule {
	return func(flag *pflag.Flag, cmd *cobra.Command) error {
		for _, word := range properNouns {
			if strings.HasPrefix(flag.Usage, word) {
				return nil
			}
		}

		if len(flag.Usage) > 0 && (flag.Usage[0] < 'A' || flag.Usage[0] > 'Z') {
			return fmt.Errorf("flag usage should start with a capital for --%s on `%s`", flag.Name, FullCommand(cmd))
		}
		return nil
	}
}

// RequireFlagNameLength checks that a flag is between a certain min and max length
func RequireFlagNameLength(minLength, maxLength int) FlagRule {
	return func(flag *pflag.Flag, cmd *cobra.Command) error {
		var issues *multierror.Error
		if len(flag.Name) < minLength {
			issue := fmt.Errorf("flag name is too short for --%s on `%s` (%d < %d)", flag.Name, FullCommand(cmd), len(flag.Name), minLength)
			issues = multierror.Append(issues, issue)
		}
		if len(flag.Name) > maxLength {
			issue := fmt.Errorf("flag name is too long for --%s on `%s` (%d > %d)", flag.Name, FullCommand(cmd), len(flag.Name), maxLength)
			issues = multierror.Append(issues, issue)
		}
		return issues
	}
}

// RequireFlagKebabCase checks that a flag is kebab-case
func RequireFlagKebabCase(flag *pflag.Flag, cmd *cobra.Command) error {
	flagKebab := strcase.ToKebab(flag.Name)
	if flagKebab != flag.Name {
		return fmt.Errorf("flag name must be kebab-case: --%s should be --%s on `%s`", flag.Name, flagKebab, FullCommand(cmd))
	}
	return nil
}

func RequireFlagUsageRealWords(properNouns []string) FlagRule {
	return func(flag *pflag.Flag, cmd *cobra.Command) error {
		var issues *multierror.Error
		usage := strings.TrimRight(alphanumeric.ReplaceAllString(flag.Usage, " "), " ") // Remove any punctuation before checking spelling
		if usage == "" {
			return nil
		}

		for _, w := range strings.Split(usage, " ") {
			if ok := vocab.Spell(w); !ok && !types.Contains(properNouns, w) {
				issue := fmt.Errorf("flag usage should consist of delimited real english words for --%s on `%s` - unknown '%s' in '%s'", flag.Name, FullCommand(cmd), w, usage)
				issues = multierror.Append(issues, issue)
			}
		}
		return issues.ErrorOrNil()
	}
}

func RequireStringSlicePrefix(flag *pflag.Flag, cmd *cobra.Command) error {
	if flag.Value.Type() == "stringSlice" && !strings.HasPrefix(flag.Usage, "A comma-separated list of") {
		return fmt.Errorf("%s: flag `--%s` usage must begin with \"A comma-separated list of\"", cmd.CommandPath(), flag.Name)
	}
	return nil
}
