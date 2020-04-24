package completer

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"testing"
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
			name:   "RootCommandTest",
			fields: fields{
				RootCmd: createNestedCommand(0, false),
			},
			levels: 0,
			args:   args{
						d: *prompt.NewDocument(),
			},
			want:   []prompt.Suggest{},
		},
		{
			name:   "SimpleCommandTest",
			fields: fields{
				RootCmd: createNestedCommand(1, false),
			},
			levels: 1,
			args:   args{
				d: prompt.Document{Text: ""},
			},
			want:   expectedSuggestions(1, []string{"a", "b", "c"}),
		},
		{
			name:   "PartialCommandTest",
			fields: fields{
				RootCmd: createNestedCommand(1, false),
			},
			levels: 1,
			args:   args{
				d: *createDocument("level-a"),
			},
			want:   expectedSuggestions(1, []string{"a"}),
		},
		{
			name:   "FlagCommandTest",
			fields: fields{
				RootCmd: createNestedCommand(1, true),
			},
			levels: 1,
			args:   args{
				d: *createDocument("--"),
			},
			want:   []prompt.Suggest{
					{
						Text: "--flag",
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
			got := c.Complete(tt.args.d);
			if len(got) != len(tt.want) {
				t.Errorf("Complete() = %v, want %v", got, tt.want)
			} else {
				for i, suggest := range got {
					expected := tt.want[i]
					if suggest != expected {
						t.Errorf("Complete() = %v, want %v", got, tt.want)
						break
					}
				}
			}
		})
	}
}

func createDocument(s string) *prompt.Document {
	buf:= prompt.NewBuffer()
	buf.InsertText(s, false, true)

	return buf.Document()
}

func expectedSuggestions(level int, subcommands []string) []prompt.Suggest {
	var expected []prompt.Suggest
	for _, s := range subcommands {
		expected = append(expected, prompt.Suggest{
			Text:        fmt.Sprintf("level-%s-%d", s, level),
			Description: fmt.Sprintf("this is command %s on level %d of the root command", s, level),
		})
	}

	return expected
}

func createNestedCommand(levels int, flags bool) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:                        "root",
		Short:                      "this is the root command at level 0",
	}

	if flags {
		rootCmd.Flags().String("flag", "default", "Just a flag")
		rootCmd.Flags().SortFlags = false
	}

	addNestedCommands(rootCmd, levels)
	return rootCmd
}

func addNestedCommands(cmd *cobra.Command, levels int) {
	if levels < 1 {
		return
	}

	for _, s := range []string{"a", "b", "c"} {
		subCmd := &cobra.Command{
			Use:	fmt.Sprintf("level-%s-%d", s, levels),
			Short:	fmt.Sprintf("this is command %s on level %d of the root command", s, levels),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(cmd.Use)
			},
		}
		addNestedCommands(subCmd, levels - 1)
		cmd.AddCommand(subCmd)
	}
}
