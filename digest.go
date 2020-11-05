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
	FORMAT_OPTIONS 	= []string{"error-digest", "color"}
	inFailedTest 	= false
	inPassedTest 	= false
	inNoTest		= false
	inRun			= false
	inFailedRun		= false
	buffer			= ""
)

func main() {
	format := strings.ToLower(os.Args[1])
	if !checkFormatIsValid(format) {
		fmt.Printf("Invalid FORMAT value provided. The current options are %v\n", FORMAT_OPTIONS)
		return
	}
	if format == "error-digest" {
		colorpkg.Set(colorpkg.FgRed)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		//line := strings.TrimSpace(string(input))
		line := string(input)
		if err != nil && err == io.EOF {
			break
		}
		switch format {
		case "error-digest":
			errorDigest(line)
		case "color":
			color(line)
		}
	}
	colorpkg.Unset()
}

func checkFormatIsValid(format string) bool {
	for _, option := range FORMAT_OPTIONS {
		if option == format {
			return true
		}
	}
	return false
}

func errorDigest(line string) {
	trimmed := strings.TrimSpace(line)
	if !inRun {
		buffer = ""
	}
	switch {
	case strings.HasPrefix(trimmed, "=== RUN"):
		buffer = ""
		updateStatusVals_InRun()
		buffer += line + "\n"
	case inRun && strings.HasPrefix(trimmed, "Error Trace"):
		updateStatusVals(false, false, false, false, true)
		fmt.Printf("%s\n", buffer)
	case strings.Contains(line, "[no test files]"):
		inFailedTest = false
	case strings.HasPrefix(trimmed, "--- PASS"): // passed
		inFailedTest = false
	case strings.HasPrefix(trimmed, "ok"):
		inFailedTest = false
	case strings.HasPrefix(trimmed, "PASS"):
		fmt.Print(".")
		inFailedTest = false
	// skipped
	case strings.HasPrefix(trimmed, "--- SKIP"):
		inFailedTest = false
	// failed
	case strings.HasPrefix(trimmed, "--- FAIL"):
		inFailedTest = true
		fmt.Println()
		fmt.Printf("%s\n", line)
	case strings.HasPrefix(trimmed, "FAIL"):
		inFailedTest = true
		fmt.Println()
		fmt.Printf("%s\n", line)
	default:
		if inFailedTest || inFailedRun {
			fmt.Printf("%s\n", line)
		} else if inRun {
			buffer += line + "\n"
		}
	}
}

func color(line string) {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "=== RUN"):
		if buffer != "" {
			colorpkg.Set(colorpkg.FgCyan)
			fmt.Printf("%s\n", buffer)
			buffer = ""
		}
		updateStatusVals_InRun()
	case inRun && strings.HasPrefix(trimmed, "Error Trace"):
		updateStatusVals(false, false, false, false, true)
	case strings.Contains(trimmed, "[no test files]"):
		updateStatusVals_NoTest()
		colorpkg.Set(colorpkg.FgCyan)
	case strings.HasPrefix(trimmed, "--- PASS"): // passed
		updateStatusVals_PassedTest()
		colorpkg.Set(colorpkg.FgGreen)
		line = fmt.Sprintf("✅ %s", line)
	case strings.HasPrefix(trimmed, "ok"):
		updateStatusVals_PassedTest()
		colorpkg.Set(colorpkg.FgGreen)
		line = fmt.Sprintf("✅ %s", line)
	case strings.HasPrefix(trimmed, "PASS"):
		updateStatusVals_PassedTest()
		colorpkg.Set(colorpkg.FgGreen)
		line = fmt.Sprintf("✅ %s", line)
	// skipped
	case strings.HasPrefix(trimmed, "--- SKIP"):
		updateStatusVals_NoTest()
		colorpkg.Set(colorpkg.FgCyan)
	// failed
	case strings.HasPrefix(trimmed, "--- FAIL"):
		updateStatusVals_FailedTest()
		colorpkg.Set(colorpkg.FgRed)
		line = fmt.Sprintf("❌ %s", line)
	case strings.HasPrefix(trimmed, "FAIL"):
		updateStatusVals_FailedTest()
		colorpkg.Set(colorpkg.FgRed)
		line = fmt.Sprintf("❌ %s", line)
	default:
		if inFailedTest || inFailedRun {
			colorpkg.Set(colorpkg.FgRed)
		} else if inPassedTest {
			colorpkg.Set(colorpkg.FgGreen)
		} else if inNoTest {
			colorpkg.Set(colorpkg.FgCyan)
		}
	}
	if !inRun {
		if buffer != "" {
			if inFailedRun {
				colorpkg.Set(colorpkg.FgRed)
			} else {
				colorpkg.Set(colorpkg.FgCyan)
			}
			fmt.Printf("%s\n", buffer)
			buffer = ""
		}
		fmt.Printf("%s\n", line)
	} else {
		buffer += line + "\n"
	}
	colorpkg.Unset()
}

func updateStatusVals(failedTest bool, passedTest bool, noTest bool, run bool, failedRun bool) {
	inFailedTest	= failedTest
	inPassedTest	= passedTest
	inNoTest		= noTest
	inRun			= run
	inFailedRun		= failedRun
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
