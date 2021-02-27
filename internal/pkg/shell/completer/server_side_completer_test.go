package completer

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	completeChild   = "completechild"
	noCompleteChild = "nocompletechild"
	parent          = "parent"
	root            = "root"

	completeFlag            = "completeflag"
	completeFlagShorthand   = "c"
	noCompleteFlag          = "nocompleteflag"
	noCompleteFlagShorthand = "m"
	countFlag               = "count"
	countFlagShorthand      = "t"
	boolFlag                = "bool"
	boolFlagShorthand       = "b"
	staticFlag              = "staticflag"
	staticFlagShorthand     = "s"
)

var (
	expectedCommandSuggestions = []prompt.Suggest{newSuggestion("command suggestion")}
	expectedFlagSuggestions    = []prompt.Suggest{newSuggestion("flag suggestion")}
	staticFlagSuggestions      = []prompt.Suggest{newSuggestion("static flag suggestion")}
	parentInputString          = parent
	completeChildInputString   = fmt.Sprintf("%s %s", parent, completeChild)
	noCompleteChildInputString = fmt.Sprintf("%s %s", parent, noCompleteChild)
)

type testCase struct {
	name                string
	inputString         string
	expectedSuggestions []prompt.Suggest
}

type parentCompletableCommand struct {
	*cobra.Command
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
}

func (c *parentCompletableCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *parentCompletableCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *parentCompletableCommand) ServerComplete() []prompt.Suggest {
	return expectedCommandSuggestions
}

func (c *parentCompletableCommand) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return c.completableFlagChildren
}

func (c *parentCompletableCommand) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		completeFlag: func() []prompt.Suggest {
			return expectedFlagSuggestions
		},
	}
}

func prerunLoggedIn(cmd *cobra.Command, arg []string) error {
	return nil
}

func prerunNotLoggedIn(cmd *cobra.Command, arg []string) error {
	return errors.New("Not logged in")
}

type ServerSideCompleterTestSuite struct {
	suite.Suite
	serverSideCompleter ServerSideCompleter
	rootCmd             *cobra.Command
	parentCmd           *cobra.Command
	completeChild       *cobra.Command
	require             *require.Assertions
}

func (suite *ServerSideCompleterTestSuite) SetupSuite() {
	suite.require = require.New(suite.T())
}

func (suite *ServerSideCompleterTestSuite) SetupTest() {
	suite.initializeCommands()
	suite.serverSideCompleter = NewServerSideCompleter(suite.rootCmd)
	suite.addCompletableCommands()
	suite.triggerParentCommandToUpdateCache()
}

func (suite *ServerSideCompleterTestSuite) initializeCommands() {
	rootCmd := &cobra.Command{
		Use: "root",
	}
	rootCmd.PersistentFlags().BoolP(boolFlag, boolFlagShorthand, false, "bool flag")
	rootCmd.PersistentFlags().CountP(countFlag, countFlagShorthand, "count flag")

	parentCmd := &cobra.Command{
		Use: "parent",
	}
	parentCmd.PersistentPreRunE = prerunLoggedIn

	completeChild := &cobra.Command{
		Use:  completeChild,
		Args: cobra.ExactArgs(1),
	}
	completeChild.Flags().StringP(completeFlag, completeFlagShorthand, "", "Flag that has completion")
	completeChild.Flags().StringP(noCompleteFlag, noCompleteFlagShorthand, "", "Flag that does not have completion")
	completeChild.Flags().StringP(staticFlag, staticFlagShorthand, "", "Flag that has static completion")

	noCompleteChild := &cobra.Command{
		Use:  noCompleteChild,
		Args: cobra.ExactArgs(1),
	}
	noCompleteChild.Flags().StringP(completeFlag, completeFlagShorthand, "", "This flag has no completion for this command")
	noCompleteChild.Flags().StringP(noCompleteFlag, noCompleteFlagShorthand, "", "Flag that does not have completion")
	noCompleteChild.Flags().StringP(staticFlag, staticFlagShorthand, "", "This flag that has no static completion for this command")

	rootCmd.AddCommand(parentCmd)
	parentCmd.AddCommand(completeChild)
	parentCmd.AddCommand(noCompleteChild)

	suite.rootCmd = rootCmd
	suite.parentCmd = parentCmd
	suite.completeChild = completeChild
}

// only completeChild has completion
// only completeFlag and staticFlag with completeChild has completion
// noCompleteChild does not have completion for the command itself nor any of its flags
func (suite *ServerSideCompleterTestSuite) addCompletableCommands() {
	cc := &parentCompletableCommand{}
	cc.Command = suite.parentCmd
	cc.completableChildren = []*cobra.Command{suite.completeChild}
	cc.completableFlagChildren = map[string][]*cobra.Command{
		completeFlag: {suite.completeChild},
	}
	suite.serverSideCompleter.AddCommand(cc)
	// add static flag completion for completeChild command
	suite.serverSideCompleter.AddStaticFlagCompletion(staticFlag, staticFlagSuggestions, []string{fmt.Sprintf("%s %s %s", root, parent, completeChild)})
}

func (suite *ServerSideCompleterTestSuite) triggerParentCommandToUpdateCache() {
	// no suggestions expected when on parent command
	suggestions := suite.serverSideCompleter.Complete(createDocument(parentInputString + " "))
	suite.require.Empty(suggestions)
	time.Sleep(100 * time.Millisecond) // let goroutine run to update cache for child command
}

func (suite *ServerSideCompleterTestSuite) TestCommandSuggestions() {
	testCases := []testCase{
		{
			name:        "no suggestions without space after command",
			inputString: completeChildInputString,
		},
		{
			name:                "expect suggestions when command can take argument",
			inputString:         completeChildInputString + " ",
			expectedSuggestions: expectedCommandSuggestions,
		},
		{
			name:        "no suggestions if exceed number of argument",
			inputString: strings.Join([]string{completeChildInputString, "arg", ""}, " "),
		},
		{
			name:                "suggestions after flag with value",
			inputString:         strings.Join([]string{completeChildInputString, "--" + noCompleteFlag, "flagVal", ""}, " "),
			expectedSuggestions: expectedCommandSuggestions,
		},
		{
			name:                "suggestions after flag with no value",
			inputString:         strings.Join([]string{completeChildInputString, "-" + boolFlagShorthand, ""}, " "),
			expectedSuggestions: expectedCommandSuggestions,
		},
		{
			name:        "no suggestions when expecting flag value",
			inputString: strings.Join([]string{completeChildInputString, "--" + noCompleteFlag, ""}, " "),
		},
		{
			name:        "no suggestions for command not part of the completable children",
			inputString: noCompleteChildInputString + " ",
		},
		{
			name:        "no suggetions for non existant subcommand",
			inputString: strings.Join([]string{parentInputString, "notacommand", ""}, " "),
		},
	}
	suite.runTestCases(testCases)
}

func (suite *ServerSideCompleterTestSuite) TestFlagArgValidation() {
	testCases := []testCase{
		{
			name:        "no suggestions without space after command",
			inputString: strings.Join([]string{completeChildInputString, "--" + noCompleteFlag}, " "),
		},
		{
			name:                "suggestions for flag",
			inputString:         strings.Join([]string{completeChildInputString, "--" + completeFlag, ""}, " "),
			expectedSuggestions: expectedFlagSuggestions,
		},
		{
			name:                "suggestions for flag after arg",
			inputString:         strings.Join([]string{completeChildInputString, "arg", "--" + completeFlag, ""}, " "),
			expectedSuggestions: expectedFlagSuggestions,
		},
		{
			name:                "suggestions for flag after another flag",
			inputString:         strings.Join([]string{completeChildInputString, "--" + noCompleteFlag, "flagarg", "--" + completeFlag, ""}, " "),
			expectedSuggestions: expectedFlagSuggestions,
		},
		{
			name:                "suggestions for shorthand flag",
			inputString:         strings.Join([]string{completeChildInputString, "-" + completeFlagShorthand, ""}, " "),
			expectedSuggestions: expectedFlagSuggestions,
		},
		{
			name:        "no suggestions for non existent flag",
			inputString: strings.Join([]string{completeChildInputString, "--notaflag", ""}, " "),
		},
		{
			name:        "no suggestions for flag that expects no completion",
			inputString: strings.Join([]string{completeChildInputString, "--" + noCompleteFlag, ""}, " "),
		},
		{
			name:        "no suggestions for shorthand flag",
			inputString: strings.Join([]string{noCompleteChildInputString, "--" + completeFlag, ""}, " "),
		},
		{
			name:        "no suggestions when flag is mistakenly put in place of argument of another flag",
			inputString: strings.Join([]string{noCompleteChildInputString, "--" + noCompleteFlag, "--" + completeFlag}, " "),
		},
	}
	suite.runTestCases(testCases)
}

func (suite *ServerSideCompleterTestSuite) TestLoggedOutState() {
	testCases := []testCase{
		{
			name:                "expect suggestions when command can take argument",
			inputString:         completeChildInputString + " ",
			expectedSuggestions: expectedCommandSuggestions,
		},
		{
			name:                "suggestions for flag",
			inputString:         strings.Join([]string{completeChildInputString, "--" + completeFlag, ""}, " "),
			expectedSuggestions: expectedFlagSuggestions,
		},
	}
	suite.runTestCases(testCases)

	// simulate logging out by having the prerun throw error
	// the rerun to check that the cache is cleared so that no suggestions are shown
	suite.parentCmd.PersistentPreRunE = prerunNotLoggedIn
	testCases = []testCase{
		{
			name:        "no suggestions if logged out",
			inputString: completeChildInputString + " ",
		},
		{
			name:        "no suggestions if logged out",
			inputString: strings.Join([]string{completeChildInputString, "--" + completeFlag, ""}, " "),
		},
	}
	suite.runTestCases(testCases)
}

func (suite *ServerSideCompleterTestSuite) TestStaticFlagCompletion() {
	testCases := []testCase{
		{
			name:                "suggestions for static flag",
			inputString:         strings.Join([]string{completeChildInputString, "--" + staticFlag, ""}, " "),
			expectedSuggestions: staticFlagSuggestions,
		},
		{
			name:                "suggestions for shorthand static flag",
			inputString:         strings.Join([]string{completeChildInputString, "-" + staticFlagShorthand, ""}, " "),
			expectedSuggestions: staticFlagSuggestions,
		},
		{
			name:        "no suggestions when the command is not part of the completable children",
			inputString: strings.Join([]string{noCompleteChildInputString, "-" + staticFlagShorthand, ""}, " "),
		},
	}
	suite.runTestCases(testCases)
}

func (suite *ServerSideCompleterTestSuite) runTestCases(testCases []testCase) {
	t := suite.T()
	req := require.New(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suggestions := suite.serverSideCompleter.Complete(createDocument(tc.inputString))
			if len(tc.expectedSuggestions) == 0 {
				req.Empty(tc.expectedSuggestions)
			} else {
				req.Equal(tc.expectedSuggestions, suggestions)
			}
		})
	}
}

func TestServerSideCompleterTestSuite(t *testing.T) {
	suite.Run(t, new(ServerSideCompleterTestSuite))
}
