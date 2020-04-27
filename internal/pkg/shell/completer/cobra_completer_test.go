package completer

import (
	"fmt"
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
		levels int
		args   args
		want   []prompt.Suggest
	}{
		{
			name: "NoSuggestions",
			fields: fields{
				RootCmd: createCommands([]string{"a", "b", "c"}),
			},
			levels: 0,
			args: args{
				d: *createDocument("this command doesn't even exist"),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "AllSuggestions",
			fields: fields{
				RootCmd: createCommands([]string{"a", "b", "c"}),
			},
			levels: 1,
			args: args{
				d: *createDocument(""),
			},
			want: expectedSuggestions([]string{"a", "b", "c"}),
		},
		{
			name: "OneSuggestion",
			fields: fields{
				RootCmd: createCommands([]string{"aa", "bb", "cc"}),
			},
			levels: 1,
			args: args{
				d: *createDocument("b"),
			},
			want: expectedSuggestions([]string{"bb"}),
		},
		{
			name: "SomeSuggestions",
			fields: fields{
				RootCmd: createCommands([]string{"ab", "abc", "bc"}),
			},
			levels: 1,
			args: args{
				d: *createDocument("a"),
			},
			want: expectedSuggestions([]string{"ab", "abc"}),
		},
		{
			name: "NoHiddenCommandsSuggested",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createCommands([]string{"a", "b", "c"}) // No hidden param.
					for _, subcmd := range cmd.Commands() {
						subcmd.Hidden = true
					}
					return cmd
				}(),
			},
			levels: 1,
			args: args{
				d: *createDocument(""),
			},
			want: expectedSuggestions([]string{}),
		},
		{
			name: "FlagSuggestions",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createCommands([]string{"a", "b", "c"}) // No hidden param.
					cmd.Flags().String("flag", "default", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			levels: 1,
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

func expectedSuggestions(subcommands []string) []prompt.Suggest {
	var expected []prompt.Suggest
	for _, s := range subcommands {
		expected = append(expected, prompt.Suggest{
			Text:        s,
			Description: s,
		})
	}

	return expected
}

func createCommands(subcommands []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "root",
		Short: "this is the root command at level 0",
	}

	for _, s := range subcommands {
		subCmd := &cobra.Command{
			Use:   s,
			Short: s,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(cmd.Use)
			},
		}
		rootCmd.AddCommand(subCmd)
	}

	return rootCmd
}
