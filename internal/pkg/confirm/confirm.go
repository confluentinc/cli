package confirm

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func Do(out io.Writer, in io.Reader, msg string) (bool, error) {
	reader := bufio.NewReader(in)

	for {
		fmt.Fprintf(out, "%s (y/n): ", msg)

		input, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		choice := strings.TrimRightFunc(input, unicode.IsSpace)

		switch choice {
		case "yes", "y", "Y":
			return true, nil
		case "no", "n", "N":
			return false, nil
		default:
			fmt.Fprintf(out, "%s is not a valid choice\n", choice)
			continue
		}
	}
}
