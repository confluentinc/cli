// This is a hardcoded set of "linters" defining the CLI specification
// TODO: Would be much better if we could define this as a JSON "spec" or hardcoding like this
package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

var (
	debug = flag.Bool("debug", false, "print debug output")
)

func main() {
	flag.Parse()

	for _, cliName := range []string{"confluent", "ccloud"} {
		confluent, err := cmd.NewConfluentCommand(cliName, &config.Config{}, &version.Version{}, log.New())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = lint(confluent)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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

	// The "leaf" commands should generally all require an ID/name (like kafka cluster ID, api-key name, etc)
	if !cmd.HasAvailableSubCommands() {
		// skip special utility commands
		if cmd.Use != "login" && cmd.Use != "logout" &&
			cmd.Use != "version" && cmd.Use != "completion SHELL" &&
			!(cmd.Use == "update" && !cmd.Parent().HasParent()) {

			// skip resource container commands
			if cmd.Use != "list" && cmd.Use != "auth" &&
				// skip ACLs which don't have an identity (value objects rather than entities)
				!strings.Contains(fullCommand(cmd), "kafka acl") &&
				// skip api-key create since you don't get to choose a name for API keys
				!strings.Contains(fullCommand(cmd), "api-key create") {

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
					if !strings.HasSuffix(cmd.Use, "KEY") {
						issue := fmt.Errorf("bad usage string: must have KEY in %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
				} else {
					if !strings.HasSuffix(cmd.Use, "ID") && !strings.HasSuffix(cmd.Use, "NAME") {
						issue := fmt.Errorf("bad usage string: must have ID or NAME in %s", fullCommand(cmd))
						issues = multierror.Append(issues, issue)
					}
				}
			}

			// check whether --cluster override flag is available
			if cmd.Parent().Use != "environment" && cmd.Parent().Use != "service-account" &&
				// these all require explicit cluster as id/name args
				!strings.Contains(fullCommand(cmd), "kafka cluster") &&
				// this doesn't need a --cluster override since you provide the api key itself to identify it
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

			// check that flags aren't auto sorted
			if cmd.Flags().HasFlags() && cmd.Flags().SortFlags == true {
				issue := fmt.Errorf("flags unexpectedly sorted on %s", fullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}

			// check that help messages are consistent
			if len(cmd.Short) > 43 {
				issue := fmt.Errorf("short description is too long on %s", fullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}
			if cmd.Short[0] < 'A' || cmd.Short[0] > 'Z' {
				issue := fmt.Errorf("short description should start with a capital on %s", fullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}
			if cmd.Short[len(cmd.Short)-1] == '.' {
				issue := fmt.Errorf("short description ends with punctuation on %s", fullCommand(cmd))
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
			if strings.Contains(cmd.Long, "kafka") {
				issue := fmt.Errorf("long description should capitalize Kafka on %s", fullCommand(cmd))
				issues = multierror.Append(issues, issue)
			}
			// TODO: this is an _awful_ IsTitleCase heuristic
			if words := strings.Split(cmd.Short, " "); len(words) > 1 {
				for _, word := range words[1:] {
					if word[0] >= 'A' && word[0] <= 'Z' &&
						word != "Kafka" && word != "API" && word != "ACL" && word != "ACLs" && word != "ALL" {
						issue := fmt.Errorf("don't title case short description on %s - %s", fullCommand(cmd), cmd.Short)
						issues = multierror.Append(issues, issue)
					}
				}
			}
		}
	}
	return issues
}

func fullCommand(cmd *cobra.Command) string {
	use := []string{cmd.Use}
	cmd.VisitParents(func(command *cobra.Command) {
		use = append([]string{command.Use}, use...)
	})
	return strings.Join(use, " ")
}
