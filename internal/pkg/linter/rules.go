package linter

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/client9/gospell"
	"github.com/gobuffalo/flect"
	"github.com/hashicorp/go-multierror"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type (
	CommandRule func(cmd *cobra.Command) error
	FlagRule    func(flag *pflag.Flag, cmd *cobra.Command) error
)

var alphanumeric, _ = regexp.Compile("[^a-zA-Z0-9]+")

var vocab *gospell.GoSpell

func SetVocab(v *gospell.GoSpell) {
	vocab = v
}

// RequireRealWords checks that a field uses delimited-real-words, not smushcasecommands
func RequireRealWords(field string, delimiter rune) CommandRule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		var issues *multierror.Error
		bareCmd := strings.Split(fieldValue, " ")[0] // TODO should we check all parts?
		for _, w := range strings.Split(bareCmd, string(delimiter)) {
			if ok := vocab.Spell(w); !ok {
				issue := fmt.Errorf("%s should consist of delimited real english words for %s on `%s` - unknown %s", normalizeDesc(field), bareCmd, FullCommand(cmd), w)
				issues = multierror.Append(issues, issue)
			}
		}
		return issues
	}
}

// RequireEndWithPunctuation checks that a field ends with a period
func RequireEndWithPunctuation(field string, ignoreIfEndsWithCodeBlock bool) CommandRule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		chomped := strings.TrimRight(fieldValue, "\n")
		lines := strings.Split(fieldValue, "\n")
		if fieldValue != "" && chomped[len(chomped)-1] != '.' {
			lastLine := len(lines) - 1
			if lines[len(lines)-1] == "" {
				lastLine = len(lines) - 2
			}
			// ignore rule if last line is code block
			if !strings.HasPrefix(lines[lastLine], "  ") || !ignoreIfEndsWithCodeBlock {
				return fmt.Errorf("%s should end with punctuation on `%s`", normalizeDesc(field), FullCommand(cmd))
			}
		}
		return nil
	}
}

// RequireStartWithCapital checks that a field starts with a capital letter
func RequireStartWithCapital(field string) CommandRule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		if fieldValue != "" && (fieldValue[0] < 'A' || fieldValue[0] > 'Z') {
			return fmt.Errorf("%s should start with a capital on `%s`", normalizeDesc(field), FullCommand(cmd))
		}
		return nil
	}
}

// RequireCapitalizeProperNouns checks that a field capitalizes proper nouns
func RequireCapitalizeProperNouns(field string, properNouns []string) CommandRule {
	index := map[string]string{}
	for _, n := range properNouns {
		index[strings.ToLower(n)] = n
	}
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		var issues *multierror.Error
		for _, word := range strings.Split(fieldValue, " ") {
			if v, found := index[strings.ToLower(word)]; found && word != v {
				issue := fmt.Errorf("%s should capitalize %s on `%s`", normalizeDesc(field), v, FullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}
		}
		return issues.ErrorOrNil()
	}
}

// RequireLengthBetween checks that a field is between a certain min and max length
func RequireLengthBetween(field string, minLength, maxLength int) CommandRule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		var issues *multierror.Error
		if len(fieldValue) < minLength {
			issue := fmt.Errorf("%s is too short on `%s` (%d < %d)", normalizeDesc(field), FullCommand(cmd), len(fieldValue), minLength)
			issues = multierror.Append(issues, issue)
		}
		if len(fieldValue) > maxLength {
			issue := fmt.Errorf("%s is too long on `%s` (%d > %d)", normalizeDesc(field), FullCommand(cmd), len(fieldValue), maxLength)
			issues = multierror.Append(issues, issue)
		}
		return issues
	}
}

// RequireSingular checks that a field is singular (not plural)
func RequireSingular(field string) CommandRule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		if flect.Singularize(fieldValue) != fieldValue {
			return fmt.Errorf("%s should be singular for `%s`", normalizeDesc(field), FullCommand(cmd))
		}
		return nil
	}
}

// RequireLowerCase checks that a field is lower case
func RequireLowerCase(field string) CommandRule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		command := strings.Split(fieldValue, " ")[0]
		if strings.ToLower(command) != command {
			return fmt.Errorf("%s should be lower case for `%s`", normalizeDesc(field), FullCommand(cmd))
		}
		return nil
	}
}

// NamedArgumentConfig lets you specify different argument names in the help/usage
// for create commands vs other commands; e.g., to pass NAME on create and ID elsewhere.
type NamedArgumentConfig struct {
	CreateCommandArg string
	OtherCommandsArg string
}

func isCapitalized(word string) bool {
	return word[0] >= 'A' && word[0] <= 'Z'
}

func requireNotTitleCaseHelper(fieldValue string, properNouns []string, field string, fullCommand string) *multierror.Error {
	var issues *multierror.Error

	fieldValue = strings.TrimSuffix(fieldValue, ".")
	for _, properNoun := range properNouns {
		fieldValue = strings.ReplaceAll(fieldValue, properNoun, "")
	}

	words := strings.Split(fieldValue, " ")
	for i := 0; i < len(words); i++ {
		word := strings.TrimRight(alphanumeric.ReplaceAllString(words[i], ""), " ") // Remove any punctuation before comparison
		if word == "" {
			continue
		}
		if i == 0 {
			if isCapitalized(word) {
				continue
			} else {
				issue := fmt.Errorf("should capitalize %s on `%s` - %s", normalizeDesc(field), fullCommand, fieldValue)
				issues = multierror.Append(issues, issue)
			}
		}
		if !isCapitalized(word) {
			continue
		}
		// Starting a new sentence
		if i > 0 && strings.HasSuffix(words[i-1], ".") {
			continue
		}
		issue := fmt.Errorf("don't title case %s on `%s` - %s", normalizeDesc(field), fullCommand, fieldValue)
		issues = multierror.Append(issues, issue)
	}
	return issues
}

// RequireNotTitleCase checks that a field is Not Title Casing Everything.
// You may pass a list of proper nouns that should always be capitalized, however.
func RequireNotTitleCase(field string, properNouns []string) CommandRule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field)
		return requireNotTitleCaseHelper(fieldValue, properNouns, field, FullCommand(cmd))
	}
}

func RequireListRequiredFlagsFirst() CommandRule {
	return func(cmd *cobra.Command) error {
		hasVisitedAnOptionalFlag := false
		errs := new(multierror.Error)

		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if pcmd.IsFlagRequired(flag) {
				if hasVisitedAnOptionalFlag {
					errs = multierror.Append(errs, fmt.Errorf("%s: required flag `--%s` must be listed before non-required flags", cmd.CommandPath(), flag.Name))
				}
			} else {
				hasVisitedAnOptionalFlag = true
			}
		})

		return errs
	}
}

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

// RequireFlagUsageStartWithCapital checks that a flag description starts with a capital letter
func RequireFlagUsageStartWithCapital(properNouns []string) FlagRule {
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

func RequireFlagUsageRealWords(flag *pflag.Flag, cmd *cobra.Command) error {
	var issues *multierror.Error
	usage := strings.TrimRight(alphanumeric.ReplaceAllString(flag.Usage, " "), " ") // Remove any punctuation before checking spelling
	if usage == "" {
		return nil
	}
	for _, w := range strings.Split(usage, " ") {
		if ok := vocab.Spell(w); !ok {
			issue := fmt.Errorf("flag usage should consist of delimited real english words for --%s on `%s` - unknown '%s' in '%s'", flag.Name, FullCommand(cmd), w, usage)
			issues = multierror.Append(issues, issue)
		}
	}
	return issues.ErrorOrNil()
}

func getValueByName(obj interface{}, name string) string {
	return reflect.Indirect(reflect.ValueOf(obj)).FieldByName(name).String()
}

func normalizeDesc(field string) string {
	switch field {
	case "Use":
		return "command"
	case "Long":
		return "long description"
	case "Short":
		return "short description"
	default:
		return strings.ToLower(field)
	}
}
