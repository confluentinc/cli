package controller

import (
	"os"

	"github.com/fatih/color"
	"golang.org/x/term"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func getConsoleParser() prompt.ConsoleParser {
	consoleParser := prompt.NewStandardInputParser()
	err := consoleParser.Setup()
	if err != nil {
		log.CliLogger.Warnf("Couldn't setup console parser. Error: %v\n", err)
	}
	return consoleParser
}

func tearDownConsoleParser(consoleParser prompt.ConsoleParser) {
	err := consoleParser.TearDown()
	if err != nil {
		log.CliLogger.Warnf("Couldn't tear down console parser. Error: %v\n", err)
	}
}

func getStdin() *term.State {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.CliLogger.Warnf("Couldn't get stdin state with term.GetState. Error: %v\n", err)
		return nil
	}
	return state
}

func restoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}

func outputErr(s string) {
	c := color.New(config.ErrorColor)
	output.Println(c.Sprintf(s))
}

func outputErrf(s string, args ...any) {
	c := color.New(config.ErrorColor)
	output.Printf(c.Sprint(s), args...)
}

func outputInfo(s string) {
	c := color.New(config.InfoColor)
	output.Println(c.Sprint(s))
}

func outputInfof(s string, args ...any) {
	c := color.New(config.InfoColor)
	output.Printf(c.Sprint(s), args...)
}

func outputWarn(s string) {
	c := color.New(config.WarnColor)
	output.Println(c.Sprint(s))
}
