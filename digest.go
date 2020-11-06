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
	currColor		= colorpkg.FgRed
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
		//switch format {
		//case "error-digest":
		//	errorDigest(line)
		//case "color":
		//	color(line)
		//}
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
	case strings.HasPrefix(trimmed, "--- PASS") || strings.HasPrefix(trimmed, "PASS") || strings.HasPrefix(trimmed, "ok"): // passed
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
	if currColor == colorpkg.FgRed {
		currColor = colorpkg.FgMagenta
		return currColor
	} else {
		currColor = colorpkg.FgRed
		return currColor
	}
}

//func color2(line string) {
//	trimmed := strings.TrimSpace(line)
//	switch {
//	case strings.HasPrefix(trimmed, "=== RUN"):
//		writeOrClearErrDigestBuffer()
//		updateStatusVals_InRun()
//	case inRun && strings.HasPrefix(trimmed, "Error Trace"):
//		updateStatusVals_FailedRun()
//	case strings.HasPrefix(trimmed, "--- PASS") || strings.HasPrefix(trimmed, "PASS") || strings.HasPrefix(trimmed, "ok"): // passed
//		writeOrClearErrDigestBuffer()
//		updateStatusVals_PassedTest()
//	// skipped
//	case strings.HasPrefix(trimmed, "--- SKIP") || strings.Contains(line, "[no test files]"):
//		writeOrClearErrDigestBuffer()
//		updateStatusVals_NoTest()
//	// failed
//	case strings.HasPrefix(trimmed, "--- FAIL") || strings.HasPrefix(trimmed, "FAIL"):
//		writeOrClearErrDigestBuffer()
//		updateStatusVals_FailedTest()
//	}
//	buffer += line + "\n"
//}
//
//func color(line string) {
//	trimmed := strings.TrimSpace(line)
//	switch {
//	case strings.HasPrefix(trimmed, "=== RUN"):
//		if buffer != "" {
//			colorpkg.Set(colorpkg.FgCyan)
//			fmt.Printf("%s\n", buffer)
//			buffer = ""
//		}
//		updateStatusVals_InRun()
//	case inRun && strings.HasPrefix(trimmed, "Error Trace"):
//		updateStatusVals(false, false, false, false, true)
//	case strings.Contains(trimmed, "[no test files]"):
//		updateStatusVals_NoTest()
//		colorpkg.Set(colorpkg.FgCyan)
//	case strings.HasPrefix(trimmed, "--- PASS"): // passed
//		updateStatusVals_PassedTest()
//		colorpkg.Set(colorpkg.FgGreen)
//		line = fmt.Sprintf("✅ %s", line)
//	case strings.HasPrefix(trimmed, "ok"):
//		updateStatusVals_PassedTest()
//		colorpkg.Set(colorpkg.FgGreen)
//		line = fmt.Sprintf("✅ %s", line)
//	case strings.HasPrefix(trimmed, "PASS"):
//		updateStatusVals_PassedTest()
//		colorpkg.Set(colorpkg.FgGreen)
//		line = fmt.Sprintf("✅ %s", line)
//	// skipped
//	case strings.HasPrefix(trimmed, "--- SKIP"):
//		updateStatusVals_NoTest()
//		colorpkg.Set(colorpkg.FgCyan)
//	// failed
//	case strings.HasPrefix(trimmed, "--- FAIL"):
//		updateStatusVals_FailedTest()
//		colorpkg.Set(colorpkg.FgRed)
//		line = fmt.Sprintf("❌ %s", line)
//	case strings.HasPrefix(trimmed, "FAIL"):
//		updateStatusVals_FailedTest()
//		colorpkg.Set(colorpkg.FgRed)
//		line = fmt.Sprintf("❌ %s", line)
//	default:
//		if inFailedTest || inFailedRun {
//			colorpkg.Set(colorpkg.FgRed)
//		} else if inPassedTest {
//			colorpkg.Set(colorpkg.FgGreen)
//		} else if inNoTest {
//			colorpkg.Set(colorpkg.FgCyan)
//		}
//	}
//	if !inRun {
//		if buffer != "" {
//			if inFailedRun {
//				colorpkg.Set(colorpkg.FgRed)
//			} else {
//				colorpkg.Set(colorpkg.FgCyan)
//			}
//			fmt.Printf("%s\n", buffer)
//			buffer = ""
//		}
//		fmt.Printf("%s\n", line)
//	} else {
//		buffer += line + "\n"
//	}
//	colorpkg.Unset()
//}

//func updateStatusVals(failedTest bool, passedTest bool, noTest bool, run bool, failedRun bool) {
//	inFailedTest	= failedTest
//	inPassedTest	= passedTest
//	inNoTest		= noTest
//	inRun			= run
//	inFailedRun		= failedRun
//}

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
