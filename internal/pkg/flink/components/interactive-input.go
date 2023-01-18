package components

import (
	"fmt"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"

	"github.com/olekukonko/tablewriter"
)

func completer(in prompt.Document) []prompt.Suggest {

	s := []prompt.Suggest{
		{Text: "SELECT", Description: "Select data from a database"},
		{Text: "INSERT", Description: "Add rows to a table"},
		{Text: "DESCRIBE", Description: "Describe the schema of a table or a view"},
		{Text: "SET", Description: "Set current database or catalog"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func promptInput(value string) string {
	prompt.NewStdoutWriter().WriteRawStr("completer")

	return prompt.Input(">>> ", completer,
		prompt.OptionTitle("sql-prompt"),
		prompt.OptionInitialBufferText(value),
		prompt.OptionHistory([]string{"SELECT * FROM users;"}),
		prompt.SwitchKeyBindMode(prompt.EmacsKeyBind),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x62},
			Fn:        prompt.GoLeftWord,
		}),
		prompt.OptionAddASCIICodeBind(prompt.ASCIICodeBind{
			ASCIICode: []byte{0x1b, 0x66},
			Fn:        prompt.GoRightWord,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn:  func(b *prompt.Buffer) { os.Exit(0) },
		}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)
}

func InteractiveInput(value string) string {
	fmt.Print("Flink SQL Client \n")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;93m%s \033[0m", "[CtrlQ]", "Quit")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;93m%s \033[0m", "[CtrlS]", "Smart Completion ")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;93m%s \033[0m \n \n", "[CtrlM]", "Multiline")

	fmt.Print("flinkSQL")
	//prompt.NewStdoutWriter().WriteRawStr("testt")

	var in = promptInput(value)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"OrderDate", "Region", "Rep", "Item", "Units", "UnitCost", "Total"})

	for _, tableRow := range strings.Split(tableData, "\n") {
		row := strings.Split(tableRow, "|")
		table.Append(row)
	}

	table.Render() // Send output
	return in
}
