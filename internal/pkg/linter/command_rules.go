package linter

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/client9/gospell"
	"github.com/gobuffalo/flect"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/types"
)

type CommandRule func(cmd *cobra.Command) error

var (
	alphanumeric = regexp.MustCompile("[^a-zA-Z0-9]+")
	vocab        *gospell.GoSpell
)

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
		if strings.HasSuffix(fieldValue, "quota") {
			// flect.Singularize("xx-quota") -> xx-quotum
			// this is a known issue with the package, create an exception for this
			return nil
		}
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

func requireNotTitleCaseHelper(fieldValue string, properNouns []string, field, fullCommand string) *multierror.Error {
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
		// Word is an acronym (CLI, REST, etc.)
		if word == strings.ToUpper(word) {
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
					errs = multierror.Append(errs, fmt.Errorf("%s: required flag `--%s` must be listed before optional flags", cmd.CommandPath(), flag.Name))
				}
			} else {
				hasVisitedAnOptionalFlag = true
			}
		})

		return errs
	}
}

// RequireValidExamples checks that a command's examples have the right flags.
func RequireValidExamples() CommandRule {
	return func(cmd *cobra.Command) error {
		requiredFlags := getRequiredFlags(cmd.Flags())
		allFlags := getAllFlags(cmd.Flags())

		errs := new(multierror.Error)

		for i, example := range getExampleCodeSnippets(cmd.Example) {
			for _, flag := range requiredFlags {
				if !strings.Contains(example, flag) {
					errs = multierror.Append(errs, fmt.Errorf("%s: required flag `%s` not found in example %d", cmd.CommandPath(), flag, i+1))
				}
			}

			for _, match := range regexp.MustCompile(`--[a-z\-]+`).FindAllString(example, -1) {
				if !types.Contains(allFlags, match) {
					errs = multierror.Append(errs, fmt.Errorf("%s: unknown flag `%s` found in example %d", cmd.CommandPath(), match, i+1))
				}
			}

			for _, match := range regexp.MustCompile(`--[a-z\-]+=`).FindAllString(example, -1) {
				errs = multierror.Append(errs, fmt.Errorf("%s: flag `%s` must not use \"=\" in example %d", cmd.CommandPath(), strings.TrimRight(match, "="), i+1))
			}
		}

		return errs
	}
}

func getExampleCodeSnippets(example string) []string {
	var examples []string
	for _, row := range strings.Split(example, "\n") {
		if strings.HasPrefix(row, "  $ ") {
			examples = append(examples, strings.TrimPrefix(row, "  $ "))
		}
	}
	return examples
}

func getRequiredFlags(flags *pflag.FlagSet) []string {
	var required []string
	flags.VisitAll(func(flag *pflag.Flag) {
		if pcmd.IsFlagRequired(flag) {
			required = append(required, "--"+flag.Name)
		}
	})
	return required
}

func getAllFlags(flags *pflag.FlagSet) []string {
	var all []string
	flags.VisitAll(func(flag *pflag.Flag) {
		all = append(all, "--"+flag.Name)
	})
	return all
}

func getValueByName(obj any, name string) string {
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
