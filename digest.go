package main

import (
	"bufio"
	"fmt"
	colorpkg "github.com/fatih/color"
	"io"
	"os"
	"strings"
)
var (
	FORMAT_OPTIONS  = []string{"error-digest", "color"}
	inFailedTest    = false
	inPassedTest    = false
	inNoTest        = false
	inRun           = false
	inFailedRun     = false
	buffer          = ""
	defaultErrColor = colorpkg.FgRed
)

func main() {
	format := strings.ToLower(os.Args[1])
	if !checkFormatIsValid(format) {
		fmt.Printf("Invalid FORMAT value provided. The current options are %v\n", FORMAT_OPTIONS)
		return
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		line := string(input)
		if err != nil && err == io.EOF {
			writeOrClearBuffer(format)
			break
		}
		digest(line, format)
	}
}

func checkFormatIsValid(format string) bool {
	for _, option := range FORMAT_OPTIONS {
		if option == format {
			return true
		}
	}
	return false
}

func digest(line, format string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "=== RUN"):
		writeOrClearBuffer(format)
		updateStatusVals_InRun()
	case inRun && strings.HasPrefix(trimmed, "Error Trace"):
		updateStatusVals_FailedRun()
	//passed
	case strings.HasPrefix(trimmed, "--- PASS") || strings.HasPrefix(trimmed, "PASS") || strings.HasPrefix(trimmed, "ok"):
		writeOrClearBuffer(format)
		updateStatusVals_PassedTest()
	// skipped
	case strings.HasPrefix(trimmed, "--- SKIP") || strings.Contains(line, "[no test files]"):
		writeOrClearBuffer(format)
		updateStatusVals_NoTest()
	// failed
	case strings.HasPrefix(trimmed, "--- FAIL") || strings.HasPrefix(trimmed, "FAIL"):
		writeOrClearBuffer(format)
		updateStatusVals_FailedTest()
	}
	buffer += line + "\n"
}

func writeOrClearBuffer(format string) {
	if format == "error-digest" {
		writeOrClearErrDigestBuffer()
	} else if format == "color" {
		writeOrClearColorBuffer()
	}
}

func writeOrClearErrDigestBuffer() {
	if inFailedRun || inFailedTest {
		colorpkg.Set(getDigestErrColor())
		fmt.Printf("%s", buffer)
		colorpkg.Unset()
	}
	buffer = ""
}

func writeOrClearColorBuffer() {
	if inFailedRun || inFailedTest {
		colorpkg.Set(colorpkg.FgRed)
		buffer = "❌ " + buffer
	} else if inPassedTest {
		colorpkg.Set(colorpkg.FgGreen)
		buffer = "✅ " + buffer
	} else if inNoTest || inRun {
		colorpkg.Set(colorpkg.FgCyan)
	}
	fmt.Printf("%s", buffer)
	colorpkg.Unset()
	buffer = ""
}

func getDigestErrColor() colorpkg.Attribute{
	if defaultErrColor == colorpkg.FgRed {
		defaultErrColor = colorpkg.FgMagenta
		return defaultErrColor
	} else {
		defaultErrColor = colorpkg.FgRed
		return defaultErrColor
	}
}

func updateStatusVals_InRun() {
	inFailedTest	= false
	inPassedTest	= false
	inNoTest		= false
	inRun			= true
	inFailedRun		= false
}
func updateStatusVals_PassedTest() {
	inFailedTest	= false
	inPassedTest	= true
	inNoTest		= false
	inRun			= false
	inFailedRun		= false
}

func updateStatusVals_FailedTest() {
	inFailedTest	= true
	inPassedTest	= false
	inNoTest		= false
	inRun			= false
	inFailedRun		= false
}

func updateStatusVals_NoTest() {
	inFailedTest	= false
	inPassedTest	= false
	inNoTest		= true
	inRun			= false
	inFailedRun		= false
}

func updateStatusVals_FailedRun() {
	inFailedTest	= false
	inPassedTest	= false
	inNoTest		= false
	inRun			= false
	inFailedRun		= true
}
