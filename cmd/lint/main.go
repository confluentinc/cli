// This is a hardcoded set of "linters" defining the CLI specification
// TODO: Would be much better if we could define this as a JSON "spec" or hardcoding like this
package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/client9/gospell"
	"github.com/gobuffalo/flect"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var (
	debug    = flag.Bool("debug", false, "print debug output")
	affFile  = flag.String("aff-file", "", "hunspell .aff file")
	dicFile  = flag.String("dic-file", "", "hunspell .dic file")
	alnum, _ = regexp.Compile("[^a-zA-Z0-9]+")

	vocab *gospell.GoSpell
)

func main() {
	flag.Parse()

	var err error
	vocab, err = gospell.NewGoSpell(*affFile, *dicFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	vocab.AddWordRaw("ccloud")
	vocab.AddWordRaw("kafka")
	vocab.AddWordRaw("api")
	vocab.AddWordRaw("acl")
	vocab.AddWordRaw("url")
	vocab.AddWordRaw("config")
	vocab.AddWordRaw("multizone")
	vocab.AddWordRaw("transactional")

	var issues *multierror.Error
	for _, cliName := range []string{"confluent", "ccloud"} {
		cli, err := cmd.NewConfluentCommand(cliName, &config.Config{}, &version.Version{}, log.New())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = lint(cli)
		if err != nil {
			issues = multierror.Append(issues, err)
		}
	}
	if issues.ErrorOrNil() != nil {
		fmt.Println(issues)
		os.Exit(1)
	}
}

func lint(cmd *cobra.Command) error {
	var issues *multierror.Error

	err := linters(cmd)
	if err != nil {
		issues = multierror.Append(issues, err)
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := lint(c); err != nil {
			issues = multierror.Append(issues, err)
		}
	}

	return issues.ErrorOrNil()
}

func linters(cmd *cobra.Command) *multierror.Error {
	if *debug {
		fmt.Println(fullCommand(cmd))
		fmt.Println(cmd.Short)
		fmt.Println()
	}

	var issues *multierror.Error

	// The first set of linters only apply to "leaf" commands
	if !cmd.HasAvailableSubCommands() {

		// skip special utility commands
		if cmd.Use != "login" && cmd.Use != "logout" &&
			cmd.Use != "version" && cmd.Use != "completion SHELL" &&
			!(cmd.Use == "update" && !cmd.Parent().HasParent()) {

			// check whether an ID/name is consistently provided (like kafka cluster ID, kafka topic name, etc)
			if cmd.Use != "list" && cmd.Use != "auth" && // skip resource container commands
				// skip ACLs which don't have an identity (value objects rather than entities)
				!strings.Contains(fullCommand(cmd), "kafka acl") &&
				// skip api-key create since you don't get to choose a name for API keys
				!strings.Contains(fullCommand(cmd), "api-key create") &&
				// skip local which delegates to bash commands
				!strings.Contains(fullCommand(cmd), "local") {

				// check whether arg parsing is setup correctly
				if reflect.ValueOf(cmd.Args).Pointer() != reflect.ValueOf(cobra.ExactArgs(1)).Pointer() {
					issue := fmt.Errorf("missing expected argument on %s", fullCommand(cmd))
					issues = multierror.Append(issues, issue)
				}

				// check whether the usage string is setup correctly
				if cmd.Parent().Use == "topic" {
					if !strings.HasSuffix(cmd.Use, "TOPIC") {
						issue := fmt.Errorf("bad usage string: must have TOPIC in %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
				} else if cmd.Parent().Use == "api-key" {
					if !strings.HasSuffix(cmd.Use, "KEY") &&
						// skip for api-key store command
						!strings.HasPrefix(cmd.Use, "store") {
						issue := fmt.Errorf("bad usage string: must have KEY in %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
				} else {
					// check for "create NAME" and "<verb> ID" elsewhere
					if strings.HasPrefix(cmd.Use, "create ") {
						if !strings.HasSuffix(cmd.Use, "NAME") {
							issue := fmt.Errorf("bad usage string: must have NAME in %s", fullCommand(cmd))
							issues = multierror.Append(issues, issue)
						}
					} else if !strings.HasSuffix(cmd.Use, "ID") {
						issue := fmt.Errorf("bad usage string: must have ID in %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
				}
			}

			// check whether --cluster override flag is available
			if cmd.Parent().Use != "environment" && cmd.Parent().Use != "service-account" && cmd.Use != "local" &&
				// these all require explicit cluster as id/name args
				!strings.Contains(fullCommand(cmd), "kafka cluster") &&
				// this doesn't need a --cluster override since you provide the api key itself to identify it
				!strings.Contains(fullCommand(cmd), "api-key update") &&
				!strings.Contains(fullCommand(cmd), "api-key delete") {
				f := cmd.Flag("cluster")
				if f == nil {
					issue := fmt.Errorf("missing --cluster override flag on %s", fullCommand(cmd))
					issues = multierror.Append(issues, issue)
				} else {
					// TODO: ensuring --cluster is optional DOES NOT actually ensure that the cluster context is used
					if f.Annotations[cobra.BashCompOneRequiredFlag] != nil &&
						f.Annotations[cobra.BashCompOneRequiredFlag][0] == "true" {
						issue := fmt.Errorf("required --cluster flag should be optional on %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}

					// check that --cluster has the right type and description (so its not a different meaning)
					if f.Value.Type() != "string" {
						issue := fmt.Errorf("standard --cluster flag has the wrong type on %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
					if cmd.Parent().Use != "api-key" && f.Usage != "Kafka cluster ID" {
						issue := fmt.Errorf("bad usage string: expected standard --cluster on %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
				}
			}
		}
	}

	// check that flags aren't auto sorted
	if cmd.Flags().HasFlags() && cmd.Flags().SortFlags == true {
		issue := fmt.Errorf("flags unexpectedly sorted on %s", fullCommand(cmd))
		issues = multierror.Append(issues, issue)
	}

	// check whether commands are all lower case
	command := strings.Split(cmd.Use, " ")[0]
	if strings.ToLower(command) != command {
		issue := fmt.Errorf("commands should be lower case for %s", command)
		issues = multierror.Append(issues, issue)
	}

	// check whether resource names are singular
	if flect.Singularize(cmd.Use) != cmd.Use {
		issue := fmt.Errorf("resource names should be singular for %s", cmd.Use)
		issues = multierror.Append(issues, issue)
	}

	// check that help messages are consistent
	if len(cmd.Short) < 13 {
		issue := fmt.Errorf("short description is too short on %s - %s", fullCommand(cmd), cmd.Short)
		issues = multierror.Append(issues, issue)
	}
	if len(cmd.Short) > 55 {
		issue := fmt.Errorf("short description is too long on %s", fullCommand(cmd))
		issues = multierror.Append(issues, issue)
	}
	if cmd.Short[0] < 'A' || cmd.Short[0] > 'Z' {
		issue := fmt.Errorf("short description should start with a capital on %s", fullCommand(cmd))
		issues = multierror.Append(issues, issue)
	}
	if cmd.Short[len(cmd.Short)-1] == '.' {
		issue := fmt.Errorf("short description should not end with punctuation on %s", fullCommand(cmd))
		issues = multierror.Append(issues, issue)
	}
	if strings.Contains(cmd.Short, "kafka") {
		issue := fmt.Errorf("short description should capitalize Kafka on %s", fullCommand(cmd))
		issues = multierror.Append(issues, issue)
	}
	if cmd.Long != "" && (cmd.Long[0] < 'A' || cmd.Long[0] > 'Z') {
		issue := fmt.Errorf("long description should start with a capital on %s", fullCommand(cmd))
		issues = multierror.Append(issues, issue)
	}
	chomped := strings.TrimRight(cmd.Long, "\n")
	lines := strings.Split(cmd.Long, "\n")
	if cmd.Long != "" && chomped[len(chomped)-1] != '.' {
		lastLine := len(lines) - 1
		if lines[len(lines)-1] == "" {
			lastLine = len(lines) - 2
		}
		// ignore rule if last line is code block
		if !strings.HasPrefix(lines[lastLine], "  ") {
			issue := fmt.Errorf("long description should end with punctuation on %s", fullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
	}
	if strings.Contains(cmd.Long, "kafka") {
		issue := fmt.Errorf("long description should capitalize Kafka on %s", fullCommand(cmd))
		issues = multierror.Append(issues, issue)
	}
	// TODO: this is an _awful_ IsTitleCase heuristic
	if words := strings.Split(cmd.Short, " "); len(words) > 1 {
		for i, word := range words[1:] {
			word = alnum.ReplaceAllString(word, "") // Remove any punctuation before comparison
			if word[0] >= 'A' && word[0] <= 'Z' &&
				word != "Apache" && word != "Kafka" &&
				word != "CLI" && word != "API" && word != "ACL" && word != "ACLs" && word != "ALL" &&
				word != "Confluent" && !(words[i] == "Confluent" && word == "Cloud") && !(words[i] == "Confluent" && word == "Platform") {
				issue := fmt.Errorf("don't title case short description on %s - %s", fullCommand(cmd), cmd.Short)
				issues = multierror.Append(issues, issue)
			}
		}
	}

	// don't allow smushcasecommands, require dash-separated real words
	bareCmd := strings.Split(cmd.Use, " ")[0]
	for _, w := range strings.Split(bareCmd, "-") {
		if ok := vocab.Spell(w); !ok {
			issue := fmt.Errorf("commands should consist of dash-separated real english words for %s on %s", bareCmd, fullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
	}

	// check that flags are consistent
	cmd.Flags().VisitAll(func(pf *pflag.Flag) {
		if len(pf.Name) > 16 && pf.Name != "service-account-id" && pf.Name != "replication-factor" {
			issue := fmt.Errorf("flag name is too long for %s on %s", pf.Name, fullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
		if pf.Usage[0] < 'A' || pf.Usage[0] > 'Z' {
			issue := fmt.Errorf("flag usage should start with a capital for %s on %s", pf.Name, fullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
		if pf.Usage[len(pf.Usage)-1] == '.' {
			issue := fmt.Errorf("flag usage ends with punctuation for %s on %s", pf.Name, fullCommand(cmd))
			issues = multierror.Append(issues, issue)
		}
		nonAlpha := false
		countDashes := 0
		for _, l := range pf.Name {
			if !unicode.IsLetter(l) {
				if l == '-' {
					countDashes++
					// Even 2 is too long... service-account-id is the one exception we'll allow for now
					if countDashes > 1 && pf.Name != "service-account-id" {
						issue := fmt.Errorf("flag name must only have one dash for %s on %s", pf.Name, fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
				} else {
					if !nonAlpha {
						issue := fmt.Errorf("flag name must be letters and dash for %s on %s", pf.Name, fullCommand(cmd))
						issues = multierror.Append(issues, issue)
						nonAlpha = true
					}
				}
			}
		}
		// don't allow --smushcaseflags, require dash-separated real words
		for _, w := range strings.Split(pf.Name, "-") {
			if ok := vocab.Spell(w); !ok {
				issue := fmt.Errorf("flag name should consist of dash-separated real english words for %s on %s", pf.Name, fullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}
		}
	})
	return issues
}

func fullCommand(cmd *cobra.Command) string {
	use := []string{cmd.Use}
	cmd.VisitParents(func(command *cobra.Command) {
		use = append([]string{command.Use}, use...)
	})
	return strings.Join(use, " ")
}
