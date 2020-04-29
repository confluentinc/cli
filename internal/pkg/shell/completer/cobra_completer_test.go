package completer

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCobraCompleter_Complete(t *testing.T) {
	type fields struct {
		RootCmd *cobra.Command
	}
	type args struct {
		d prompt.Document
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []prompt.Suggest
	}{
		{
			name: "suggest no commands if documents matches nothing",
			fields: fields{
				RootCmd: createNestedCommands(1, 1),
			},
			args: args{
				d: *createDocument("this command doesn't even exist"),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "suggest all commands if document is empty",
			fields: fields{
				RootCmd: createNestedCommands(1, 2),
			},
			args: args{
				d: *createDocument(""),
			},
			want: []prompt.Suggest{
				newSuggestion("1"),
				newSuggestion("11"),
			},
		},
		{
			name: "suggest some commands if document is a partial match",
			fields: fields{
				RootCmd: createNestedCommands(1, 3),
			},
			args: args{
				d: *createDocument("11"),
			},
			want: []prompt.Suggest{
				newSuggestion("11"),
				newSuggestion("111"),
			},
		},
		{
			name: "don't suggest any hidden commands",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2)
					for _, subcmd := range cmd.Commands() {
						subcmd.Hidden = true
					}
					cmd.Commands()[0].Hidden = false // "1"
					return cmd
				}(),
			},
			args: args{
				d: *createDocument(""),
			},
			want: []prompt.Suggest{
				newSuggestion("1"),
			},
		},
		{
			name: "suggest flag with no preceding command",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Flags().String("flag", "default", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: *createDocument("--"),
			},
			want: []prompt.Suggest{
				{
					Text:        "--flag",
					Description: "Just a flag",
				},
			},
		},
		{
			name: "suggest flag with a preceding command",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Commands()[0].Flags().String("flag", "", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: *createDocument("1 --"),
			},
			want: []prompt.Suggest{
				{
					Text:        "--flag",
					Description: "Just a flag",
				},
			},
		},
		{
			name: "suggest shorthand flag with no preceding command",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Flags().StringP("flag", "f", "", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: *createDocument("-"),
			},
			want: []prompt.Suggest{
				{
					Text:        "-f",
					Description: "Just a flag",
				},
			},
		},
		{
			name: "suggest shorthand flag with a preceding command",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Commands()[1].Flags().StringP("flag", "f", "", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: *createDocument("11 -"),
			},
			want: []prompt.Suggest{
				{
					Text:        "-f",
					Description: "Just a flag",
				},
			},
		},
		{
			name: "suggest shorthand flag with a preceding nested command",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Commands()[1].Commands()[0].Flags().StringP("flag", "f", "", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: *createDocument("11 2 -"),
			},
			want: []prompt.Suggest{
				{
					Text:        "-f",
					Description: "Just a flag",
				},
			},
		},
		{
			name: "suggest nested commands",
			fields: fields{
				RootCmd: createNestedCommands(2, 2),
			},
			args: args{
				d: *createDocument("2 "),
			},
			want: []prompt.Suggest{
				{
					Text:        "1",
					Description: "1",
				},
				{
					Text:        "11",
					Description: "11",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CobraCompleter{
				RootCmd: tt.fields.RootCmd,
			}
			got := c.Complete(tt.args.d)
			require.Equal(t, tt.want, got)
		})
	}
}

func createDocument(s string) *prompt.Document {
	buf := prompt.NewBuffer()
	buf.InsertText(s, false, true)
	return buf.Document()
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
	addNestedCommands(rootCmd, levels, levels, cmdsPerLevel)
	return rootCmd
}

func newSuggestion(s string) prompt.Suggest {
	return prompt.Suggest{
		Text:        s,
		Description: s,
	}
}

func addNestedCommands(rootCmd *cobra.Command, maxLevel int, levels int, cmdsPerLevel int) {
	if levels < 1 {
		return
	}
	for i := 0; i < cmdsPerLevel; i++ {
		s := strings.Repeat(strconv.Itoa(maxLevel-levels+1), i+1)
		subCmd := &cobra.Command{
			Use:   s,
			Short: s,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(cmd.Use)
			},
		}
		addNestedCommands(subCmd, maxLevel, levels-1, cmdsPerLevel)
		rootCmd.AddCommand(subCmd)
	}
}
