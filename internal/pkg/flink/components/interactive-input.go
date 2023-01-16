package components

import (
	"fmt"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"

	"github.com/olekukonko/tablewriter"
)

func promptInput() string {
	return prompt.Input(">>> ", completer,
		prompt.OptionTitle("sql-prompt"),
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
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray))
}

func InteractiveInput() {
	fmt.Print("flinkSQL")
	prompt.NewStdoutWriter().WriteRawStr("testt")

	var in = promptInput()
	fmt.Println("Your input: " + in)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"OrderDate", "Region", "Rep", "Item", "Units", "UnitCost", "Total"})

	for _, tableRow := range strings.Split(tableData, "\n") {
		row := strings.Split(tableRow, "|")
		table.Append(row)
	}

	table.Render() // Send output
}
