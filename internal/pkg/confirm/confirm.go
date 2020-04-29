package confirm

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"
)

func Do(out io.Writer, in io.Reader, msg string) bool {
	reader := bufio.NewReader(in)

	for {
		fmt.Fprintf(out, "%s (y/n): ", msg)

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		choice := strings.TrimRightFunc(input, unicode.IsSpace)

		switch choice {
		case "yes", "y", "Y":
			return true
		case "no", "n", "N":
			return false
		default:
			fmt.Fprintf(out, "%s is not a valid choice\n", choice)
			continue
		}
	}
}
