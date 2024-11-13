package components

import (
	"strings"

	fColor "github.com/fatih/color"
	"github.com/mattn/go-runewidth"

	"github.com/confluentinc/cli/v4/pkg/color"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func PrintCompletionsState(completionsEnabled bool, maxCol int) {
	msgPrefix := "Completions are now "
	PrintOptionState(msgPrefix, completionsEnabled, maxCol)
}

func PrintDiagnosticsState(diagnosticsEnabled bool, maxCol int) {
	msg := "Diagnostics are now "
	if diagnosticsEnabled {
		msg = "You should now see the errors in your SQL statement highlighted as you type, if there are any.\n" + msg
	}
	PrintOptionState(msg, diagnosticsEnabled, maxCol)
}

func PrintOptionState(prefix string, isEnabled bool, maxCol int) {
	stateMsg := "disabled"
	if isEnabled {
		stateMsg = "enabled"
	}

	lines := strings.Split(prefix, "\n")

	output.Print(false, "\n")
	for i, line := range lines {
		if i == len(lines)-1 {
			output.Print(false, line+fColor.CyanString(stateMsg))
			line = line + stateMsg
		} else {
			output.Print(false, line)
		}

		// This prints to the console the exact amount of empty characters to fill the line might have autocompletions before
		// This is to avoid the linter to complain about not using the
		output.Println(false, strings.Repeat(" ", maxCol-runewidth.StringWidth(line)))
	}
}

func PrintWelcomeHeader(appOtions types.ApplicationOptions) {
	// Print welcome message
	output.Print(false, "Welcome! \n")
	output.Print(false, "To exit, press Ctrl-Q or type \"exit\". \n\n")

	// Print shortcuts
	c := fColor.New(color.AccentColor)

	if featureflags.Manager.BoolVariation("cli.flink.shell.enable_diagnostics_toggle", appOtions.Context, config.CliLaunchDarklyClient, true, false) {
		output.Printf(false, "[Ctrl-Q] %s [Ctrl-S] %s [Ctrl-G] %s \n", c.Sprint("Quit"), c.Sprint("Toggle Completions"), c.Sprint("Toggle Diagnostics"))
	} else {
		output.Printf(false, "[Ctrl-Q] %s [Ctrl-S] %s \n", c.Sprint("Quit"), c.Sprint("Toggle Completions"))
	}
}
