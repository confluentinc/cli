package completer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

func createDocument(s string) prompt.Document {
	buf := prompt.NewBuffer()
	buf.InsertText(s, false, true)
	return *buf.Document()
}

func createNestedCommands(levels int, cmdsPerLevel int) (cmd *cobra.Command) {
	if levels < 1 {
		return
	}
	rootCmd := &cobra.Command{
		Use:   "0",
		Short: "0",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Use)
		},
	}
	addNestedCommands(rootCmd, levels, 1, cmdsPerLevel)
	return rootCmd
}

func newSuggestion(s string) prompt.Suggest {
	return prompt.Suggest{
		Text:        s,
		Description: s,
	}
}

func addNestedCommands(rootCmd *cobra.Command, maxLevel int, levels int, cmdsPerLevel int) {
	if levels > maxLevel {
		return
	}
	for i := 0; i < cmdsPerLevel; i++ {
		s := strings.Repeat(strconv.Itoa(levels), i+1)
		subCmd := &cobra.Command{
			Use:   s,
			Short: s,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(cmd.Use)
			},
		}
		addNestedCommands(subCmd, maxLevel, levels+1, cmdsPerLevel)
		rootCmd.AddCommand(subCmd)
	}
}
