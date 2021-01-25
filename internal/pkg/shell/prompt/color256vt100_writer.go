package prompt

import (
	"fmt"
	"io"
	"strconv"

	goprompt "github.com/c-bata/go-prompt"
)

type Color256VT100Writer struct {
	goprompt.ConsoleWriter
}

// testableColor256VT100Writer implements io.Writer to make Color256VT100Writer easy to test.
// go-prompt doesn't conform to io.Writer, nor expose internal buffers which makes testing not fun.
type testableColor256VT100Writer struct {
	*Color256VT100Writer
}

// SetColor sets the text color. Color256VT100Writer will interpret each color
// as a value 0-255, as specified here: https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit.
// To prevent forking of go-prompt, this writer interprets 0 as the default color, not black.
// TODO: To use black... TBD.
func (w *Color256VT100Writer) SetColor(fg, bg goprompt.Color, bold bool) {
	if bold {
		w.setDisplayAttributes(fg, bg, goprompt.DisplayBold)
	} else {
		w.setDisplayAttributes(fg, bg, goprompt.DisplayReset)
	}
}

func NewStdoutColor256VT100Writer() *Color256VT100Writer {
	return &Color256VT100Writer{ConsoleWriter: goprompt.NewStdoutWriter()}
}

func (tw *testableColor256VT100Writer) Write(p []byte) (n int, err error) {
	tw.Color256VT100Writer.WriteRaw(p)
	return len(p), nil
}

// SetDisplayAttributes to set VT100 display attributes.
func (w *Color256VT100Writer) setDisplayAttributes(fg, bg goprompt.Color, attrs ...goprompt.DisplayAttribute) {
	writeColorString(&testableColor256VT100Writer{Color256VT100Writer: w}, fg, bg, attrs...)
}

func writeColorString(w io.Writer, fg, bg goprompt.Color, attrs ...goprompt.DisplayAttribute) {
	var err error
	write := func(p []byte) {
		if err != nil {
			return
		}
		_, err = w.Write(p)
	}
	write([]byte("\x1b["))   // Control sequence introducer.
	defer write([]byte{'m'}) // final character
	separator := []byte{';'}
	for _, a := range attrs {
		b := displayAttributeToBytes(goprompt.DisplayAttribute(a))
		write(b)
		write(separator)
	}
	// Begin writing 256 color strings.
	// Foreground.
	if fg == 0 {
		write([]byte{'3', '9'}) // Reset to default fg color.
	} else {
		write([]byte(fmt.Sprintf("38;5;%d", fg))) // 8-bit foreground escape sequence.
	}
	write(separator)
	// Background.
	if bg == 0 {
		write([]byte{'4', '9'}) // Reset to default fg color.
	} else {
		write([]byte(fmt.Sprintf("48;5;%d", bg))) // 8-bit background escape sequence.
	}
}

// displayAttributeToBytes converts a DisplayAttribute to its code in bytes.
func displayAttributeToBytes(attribute goprompt.DisplayAttribute) []byte {
	val := int(attribute)
	s := strconv.Itoa(val)
	return []byte(s)
}
