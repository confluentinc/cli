package components

import (
	"fmt"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"

	"github.com/gdamore/tcell/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/rivo/tview"
)

type ExtraSlideParams struct {
	Table *tview.Table
}

var input = tview.NewInputField().
	SetText("SELECT * FROM ORDERS;").
	SetLabel("flinkSql[yellow]>>> ").
	SetFieldBackgroundColor(tcell.ColorDefault).
	SetLabelColor(tcell.ColorWhite)

func completer(in prompt.Document) []prompt.Suggest {
	prompt.NewStdoutWriter().WriteRawStr("completer")

	s := []prompt.Suggest{
		{Text: "SELECT", Description: "Select data from a database"},
		{Text: "INSERT", Description: "Add rows to a table"},
		{Text: "DESCRIBE", Description: "Describe the schema of a table or a view"},
		{Text: "SET", Description: "Set current database or catalog"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

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

func execGoPrompt() {
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

func TableWithInput(nextSlide func(), app *tview.Application) (title string, params ExtraSlideParams, content tview.Primitive) {

	execGoPrompt()

	_, table, _, selectRow, navigate := RawTable(nextSlide, app)

	input.SetDoneFunc(func(key tcell.Key) {
		selectRow()
		navigate()
		//nextSlide()
	})

	return "Home", ExtraSlideParams{table}, tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 1, true).
		AddItem(
			(tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(table, 0, 1, true)),
			//.AddItem(list, 10, 1, false)
			0, 1, false)
}

func getInput() *tview.InputField {
	return input
}
