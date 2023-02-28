package components

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var LivePrefixState struct {
	LivePrefix string
	IsEnabled  bool
}

var LastStatement = ""
var AllStatements = ""

func Executor(in string) {
	if strings.HasSuffix(in, ";") {
		LastStatement = LastStatement + in
		LivePrefixState.IsEnabled = false
		LivePrefixState.LivePrefix = in
		AllStatements = AllStatements + LastStatement
		LastStatement = ""

		if IsInputClosingSelect(in) {
			LivePrefixState.IsEnabled = true
			LivePrefixState.LivePrefix = ""
		}

		return
	}
	LastStatement = LastStatement + in + " "
	LivePrefixState.LivePrefix = ""
	LivePrefixState.IsEnabled = true
}

func ChangeLivePrefix() (string, bool) {
	return LivePrefixState.LivePrefix, LivePrefixState.IsEnabled
}
func IsInputClosingSelect(input string) bool {
	return strings.HasPrefix(strings.ToUpper(input), "SELECT") && input[len(input)-1] == ';'
}

func init() {
	// Print Flink's ASCII Art
	b, err := os.ReadFile("components/flink_ascii_60_with_text.txt")
	if err != nil {
		log.Printf("Couldn't read flink's ascii art. Error: %v\n", err)
	}

	// TODO - After setting up the event loop, we could maybe use tcell to get the
	// terminal's width so we disable printing the ascii art if the terminal is too small
	/* screen, _ := tcell.NewScreen()
	   screen.Init()

	   w, h := screen.Size() */
	// Right now, go-prompt get's executed first so this isn't possible
	// But we can always just delete the ascii art - it's just a gimmick

	fmt.Println(string(b))

	// Print welcome message
	fmt.Fprintf(os.Stdout, "Welcome! \033[0m%s \033[0;36m%s. \033[0m \n \n", "Flink SQL Client powered", "by Confluent")

	// Print shortcuts
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlQ]", "Quit")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m", "[CtrlS]", "Smart Completion ")
	fmt.Fprintf(os.Stdout, "\033[0m%s \033[0;36m%s \033[0m \n", "[CtrlO]", "Interactive Output ON/OFF")
}
