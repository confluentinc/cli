package completer

import (
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
			name: "suggest no commands if document matches nothing",
			fields: fields{
				RootCmd: createNestedCommands(1, 1),
			},
			args: args{
				d: createDocument("this command doesn't even exist "),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "suggest all commands if document is empty",
			fields: fields{
				RootCmd: createNestedCommands(1, 2),
			},
			args: args{
				d: createDocument(""),
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
				d: createDocument("11"),
			},
			want: []prompt.Suggest{
				newSuggestion("11"),
				newSuggestion("111"),
			},
		},
		{
			name: "suggest command if document is a partial match",
			fields: fields{
				RootCmd: func() *cobra.Command {
					rootCmd := &cobra.Command{
						Use:   "a",
						Short: "a",
						Run: func(cmd *cobra.Command, args []string) {
							fmt.Println(cmd.Use)
						},
					}

					rootCmd.AddCommand(
						&cobra.Command{
							Use:   "ba",
							Short: "ba",
							Run: func(cmd *cobra.Command, args []string) {
								fmt.Println(cmd.Use)
							},
						}, &cobra.Command{
							Use:   "ca",
							Short: "ca",
							Run: func(cmd *cobra.Command, args []string) {
								fmt.Println(cmd.Use)
							},
						},
					)

					return rootCmd
				}(),
			},
			args: args{
				d: createDocument("b"),
			},
			want: []prompt.Suggest{
				newSuggestion("ba"),
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
				d: createDocument(""),
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
				d: createDocument("--"),
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
				d: createDocument("1 --"),
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
				d: createDocument("-"),
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
				d: createDocument("11 -"),
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
				d: createDocument("11 2 -"),
			},
			want: []prompt.Suggest{
				{
					Text:        "-f",
					Description: "Just a flag",
				},
			},
		},
		{
			name: "suggest flag after preceding argument",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Commands()[1].Flags().StringP("flag", "f", "", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: createDocument("11 some_arg --"),
			},
			want: []prompt.Suggest{
				{
					Text:        "--flag",
					Description: "Just a flag",
				},
			},
		},
		{
			name: "does not suggest anything after invalid flag",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Commands()[1].Flags().StringP("flag", "f", "", "Just a flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: createDocument("11 -- "),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "suggest nested commands",
			fields: fields{
				RootCmd: createNestedCommands(2, 2),
			},
			args: args{
				d: createDocument("1 "),
			},
			want: []prompt.Suggest{
				newSuggestion("2"),
				newSuggestion("22"),
			},
		},
		{
			name: "does not suggest after incorrect commands",
			fields: fields{
				RootCmd: createNestedCommands(3, 3),
			},
			args: args{
				d: createDocument("1 44 "),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "does not suggest flags without -",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 1) // No hidden param.
					cmd.PersistentFlags().StringP("one", "o", "", "Just a flag")
					cmd.PersistentFlags().StringP("two", "t", "", "Just another flag")
					cmd.Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: createDocument("1 -o "),
			},
			want: []prompt.Suggest{
				newSuggestion("2"),
			},
		},
		{
			name: "does not suggest repeated short flags",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2) // No hidden param.
					cmd.Commands()[0].Flags().StringP("one", "o", "", "Just a flag")
					cmd.Commands()[0].Flags().StringP("two", "t", "", "Just another flag")
					cmd.Commands()[0].Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: createDocument("1 -o -"),
			},
			want: []prompt.Suggest{
				{
					Text:        "-t",
					Description: "Just another flag",
				},
			},
		},
		{
			name: "does not suggest repeated long flags",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 1) // No hidden param.
					cmd.Commands()[0].Flags().StringP("one", "o", "", "Just a flag")
					cmd.Commands()[0].Flags().StringP("two", "t", "", "Just another flag")
					cmd.Commands()[0].Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: createDocument("1 -o --"),
			},
			want: []prompt.Suggest{
				{
					Text:        "--two",
					Description: "Just another flag",
				},
			},
		},
	}
	cmdLine := pflag.CommandLine
	pflag.CommandLine = nil
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CobraCompleter{
				RootCmd: tt.fields.RootCmd,
			}
			got := c.Complete(tt.args.d)
			require.Equal(t, tt.want, got)
		})
	}
	pflag.CommandLine = cmdLine
}
