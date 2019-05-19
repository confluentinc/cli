package lint_cli

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/client9/gospell"
	"github.com/gobuffalo/flect"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	alnum, _ = regexp.Compile("[^a-zA-Z0-9]+")
)

type Rule func(cmd *cobra.Command) error
type FlagRule func(flag *pflag.Flag, cmd *cobra.Command) error

var vocab *gospell.GoSpell

// TODO/HACK: this is to inject a vocab "global" object for use by the rules
func SetVocab(v *gospell.GoSpell) {
	vocab = v
}

// requireRealWords checks that we don't have smushcasecommands; require dash-separated real words
func RequireRealWords(field string) Rule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		var issues *multierror.Error
		bareCmd := strings.Split(fieldValue, " ")[0]
		for _, w := range strings.Split(bareCmd, "-") {
			if ok := vocab.Spell(w); !ok {
				issue := fmt.Errorf("%s should consist of dash-separated real english words for %s on %s",
					normalizeDesc(field), bareCmd, FullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}
		}
		return issues
	}
}

func RequireEndWithPunctuation(field string, ignoreIfEndsWithCodeBlock bool) Rule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		chomped := strings.TrimRight(fieldValue, "\n")
		lines := strings.Split(fieldValue, "\n")
		if cmd.Long != "" && chomped[len(chomped)-1] != '.' {
			lastLine := len(lines) - 1
			if lines[len(lines)-1] == "" {
				lastLine = len(lines) - 2
			}
			// ignore rule if last line is code block
			if !strings.HasPrefix(lines[lastLine], "  ") || !ignoreIfEndsWithCodeBlock {
				return fmt.Errorf("%s should end with punctuation on %s", normalizeDesc(field), FullCommand(cmd))
			}
		}
		return nil
	}
}

func RequireStartWithCapital(field string) Rule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		if fieldValue != "" && (fieldValue[0] < 'A' || fieldValue[0] > 'Z') {
			return fmt.Errorf("%s should start with a capital on %s", normalizeDesc(field), FullCommand(cmd))
		}
		return nil
	}
}

func RequireCapitalizeProperNouns(field string, properNouns []string) Rule {
	// "set" for easy search of properNouns
	whitelist := map[string]string{}
	for _, n := range properNouns {
		whitelist[strings.ToLower(n)] = n
	}
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		var issues *multierror.Error
		for _, word := range strings.Split(fieldValue, " ") {
			if v, found := whitelist[word]; found {
				issue := fmt.Errorf("%s should capitalize %s on %s", normalizeDesc(field), v, FullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}
		}
		return issues.ErrorOrNil()
	}
}

func RequireNotEndWithPunctuation(field string) Rule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		if fieldValue[len(fieldValue)-1] == '.' {
			return fmt.Errorf("%s should not end with punctuation on %s", normalizeDesc(field), FullCommand(cmd))
		}
		return nil
	}
}

// requireLengthBetween check that help messages are a consistent length
func RequireLengthBetween(field string, minLength, maxLength int) Rule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		var issues *multierror.Error
		if len(fieldValue) < minLength {
			issue := fmt.Errorf("%s is too short on %s - %s", normalizeDesc(field), FullCommand(cmd), cmd.Short)
			issues = multierror.Append(issues, issue)
		}
		if len(fieldValue) > maxLength {
			issue := fmt.Errorf("%s is too long on %s", normalizeDesc(field), FullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
		return issues
	}
}

// RequireSingular checks whether command is singular
func RequireSingular(field string) Rule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		if flect.Singularize(fieldValue) != fieldValue {
			return fmt.Errorf("%s should be singular for %s", normalizeDesc(field), FullCommand(cmd))
		}
		return nil
	}
}

// RequireLowerCase checks whether commands are all lower case
func RequireLowerCase(field string) Rule {
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		command := strings.Split(fieldValue, " ")[0]
		if strings.ToLower(command) != command {
			return fmt.Errorf("%s should be lower case for %s", normalizeDesc(field), FullCommand(cmd))
		}
		return nil
	}
}

type IDNameArgumentConfig struct {
	CreateCommandArg string
	OtherCommandsArg string
}

// requireArgument checks whether an ID/name is consistently provided (like kafka cluster ID, kafka topic name, etc)
func RequireIDNameArgument(defConfig IDNameArgumentConfig, overrides map[string]IDNameArgumentConfig) Rule {
	return func(cmd *cobra.Command) error {
		// check whether arg parsing is setup correctly to expect exactly 1 arg (the ID/Name)
		if reflect.ValueOf(cmd.Args).Pointer() != reflect.ValueOf(cobra.ExactArgs(1)).Pointer() {
			return fmt.Errorf("missing expected argument on %s", FullCommand(cmd))
		}

		// check whether the usage string is setup correctly
		if o, found := overrides[cmd.Parent().Use]; found {
			if strings.HasPrefix(cmd.Use, "create ") {
				if !strings.HasSuffix(cmd.Use, o.CreateCommandArg) {
					return fmt.Errorf("bad usage string: must have %s in %s",
						o.CreateCommandArg, FullCommand(cmd))
				}
			} else if !strings.HasSuffix(cmd.Use, o.OtherCommandsArg) {
				return fmt.Errorf("bad usage string: must have %s in %s",
					o.OtherCommandsArg, FullCommand(cmd))
			}
		} else {
			// check for "create NAME" and "<verb> ID" elsewhere
			if strings.HasPrefix(cmd.Use, "create ") {
				if !strings.HasSuffix(cmd.Use, defConfig.CreateCommandArg) {
					return fmt.Errorf("bad usage string: must have %s in %s",
						defConfig.CreateCommandArg, FullCommand(cmd))
				}
			} else if !strings.HasSuffix(cmd.Use, defConfig.OtherCommandsArg) {
				return fmt.Errorf("bad usage string: must have %s in %s",
					defConfig.OtherCommandsArg, FullCommand(cmd))
			}
		}

		return nil
	}
}

func RequireNotTitleCase(field string, alwaysCapitalize []string) Rule {
	// TODO: this is an _awful_ IsTitleCase heuristic
	// "set" of alwaysCapital words for easy search; if a multi-word phrase, key is first word only.
	// since multiple phrases can start with same word, we have an []string for each phrase (hence [][]string)
	whitelist := map[string][][]string{}
	for _, word := range alwaysCapitalize {
		parts := strings.Split(word, " ")
		// we can index multiple phrases starting with the same word
		if _, ok := whitelist[parts[0]]; ok {
			whitelist[parts[0]] = append(whitelist[parts[0]], parts)
		} else {
			whitelist[parts[0]] = [][]string{parts}
		}
	}
	return func(cmd *cobra.Command) error {
		fieldValue := getValueByName(cmd, field).String()
		var issues *multierror.Error
		if words := strings.Split(fieldValue, " "); len(words) > 1 {
			for i := 0; i < len(words); i++ {
				word := alnum.ReplaceAllString(words[i], "") // Remove any punctuation before comparison
				if word[0] >= 'A' && word[0] <= 'Z' {
					isTitleCase := true
					if _, ok := whitelist[word]; ok {
						if len(whitelist[word]) == 1 {
							isTitleCase = false
						} else {
							parts := strings.Split(word, " ")
							if phrases, ok := whitelist[parts[0]]; ok {
								for _, wp := range phrases {
									allMatch := true
									for j := 0; j < len(wp); j++ {
										if words[i+j] != wp[j] {
											allMatch = false
											break
										}
									}
									if allMatch {
										isTitleCase = false
										i += len(wp) // skip the remaining words in the phrase
									}
								}
							}
						}
					}
					if i > 0 && isTitleCase {
						issue := fmt.Errorf("don't title case %s on %s - %s",
							normalizeDesc(field), FullCommand(cmd), cmd.Short)
						issues = multierror.Append(issues, issue)
					}
				}
			}
		}
		return issues
	}
}

// RequireFlag checks whether --flag flag is available
func RequireFlag(flag string, optional bool) Rule {
	return func(cmd *cobra.Command) error {
		f := cmd.Flag(flag)
		if f == nil {
			return fmt.Errorf("missing --%s flag on %s", flag, FullCommand(cmd))
		} else {
			if optional && f.Annotations[cobra.BashCompOneRequiredFlag] != nil &&
				f.Annotations[cobra.BashCompOneRequiredFlag][0] == "true" {
				return fmt.Errorf("required --%s flag should be optional on %s", flag, FullCommand(cmd))
			}
		}
		return nil
	}
}

// RequireFlagType checks that the flag, if exists, has the specified type.
// Please use RequireFlag to ensure it exists
func RequireFlagType(flag, typeName string) Rule {
	return func(cmd *cobra.Command) error {
		f := cmd.Flag(flag)
		if f != nil {
			// check that --flag has the right type (so its not a different meaning)
			if typeName != "" && f.Value.Type() != typeName {
				return fmt.Errorf("standard --%s flag has the wrong type on %s", flag, FullCommand(cmd))
			}
		}
		return nil
	}
}

// RequireFlagDescription checks that the flag, if exists, has the specified usage string.
// Please use RequireFlag to ensure it exists
func RequireFlagDescription(flag, description string) Rule {
	return func(cmd *cobra.Command) error {
		f := cmd.Flag(flag)
		if f != nil {
			// check that --flag has the standard description (so its not a different meaning)
			if description != "" && f.Usage != description {
				return fmt.Errorf("bad usage string: expected standard description for --%s on %s",
					flag, FullCommand(cmd))
			}
		}
		return nil
	}
}

// requireFlagSort checks that flags aren't auto sorted
func RequireFlagSort(sort bool) Rule {
	return func(cmd *cobra.Command) error {
		if cmd.Flags().HasFlags() && cmd.Flags().SortFlags != sort {
			if sort {
				return fmt.Errorf("flags not sorted on %s", FullCommand(cmd))
			}
			return fmt.Errorf("flags unexpectedly sorted on %s", FullCommand(cmd))
		}
		return nil
	}
}

// RequireFlagRealWords don't allow --smushcaseflags, require dash-separated real words
func RequireFlagRealWords(pf *pflag.Flag, cmd *cobra.Command) error {
	for _, w := range strings.Split(pf.Name, "-") {
		if ok := vocab.Spell(w); !ok {
			return fmt.Errorf("flag name should consist of dash-separated real english words for %s on %s",
				pf.Name, FullCommand(cmd))
		}
	}
	return nil
}

func RequireFlagDelimiter(delim rune, maxCount int) FlagRule {
	return func(pf *pflag.Flag, cmd *cobra.Command) error {
		countDelim := 0
		for _, l := range pf.Name {
			if !unicode.IsLetter(l) {
				if l == delim {
					countDelim++
					if countDelim > maxCount  {
						return fmt.Errorf("flag name must only have %d delimiter (\"%c\") for %s on %s",
							maxCount, delim, pf.Name, FullCommand(cmd))
					}
				}
			}
		}
		return nil
	}
}

func RequireFlagCharacters(delim rune) FlagRule {
	return func(pf *pflag.Flag, cmd *cobra.Command) error {
		for _, l := range pf.Name {
			if !unicode.IsLetter(l) && l != delim {
				return fmt.Errorf("flag name must be letters and delim (\"%c\") for %s on %s",
					delim, pf.Name, FullCommand(cmd))
			}
		}
		return nil
	}
}

func RequireFlagNotEndWithPunctuation(pf *pflag.Flag, cmd *cobra.Command) error {
	if pf.Usage[len(pf.Usage)-1] == '.' {
		return fmt.Errorf("flag usage ends with punctuation for %s on %s", pf.Name, FullCommand(cmd))
	}
	return nil
}

func RequireFlagStartWithCapital(pf *pflag.Flag, cmd *cobra.Command) error {
	if pf.Usage[0] < 'A' || pf.Usage[0] > 'Z' {
		return fmt.Errorf("flag usage should start with a capital for %s on %s", pf.Name, FullCommand(cmd))
	}
	return nil
}

func RequireFlagNameLength(minLength, maxLength int) FlagRule {
	return func(pf *pflag.Flag, cmd *cobra.Command) error {
		var issues *multierror.Error
		if len(pf.Name) < minLength {
			issue := fmt.Errorf("flag name is too short for %s on %s", pf.Name, FullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
		if len(pf.Name) > maxLength {
			issue := fmt.Errorf("flag name is too long for %s on %s", pf.Name, FullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
		return issues
	}
}

func getValueByName(obj interface{}, name string) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(obj)).FieldByName(name)
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
