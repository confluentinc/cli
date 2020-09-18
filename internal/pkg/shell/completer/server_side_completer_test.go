package completer

import (
	"testing"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestServerSideCompleter_Complete(t *testing.T) {
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
			name: "should not suggest for parent command",
			fields: fields{
				RootCmd: createNestedCommands(2, 2),
			},
			args: args{
				d: createDocument("1 "),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "should not suggest for unregistered sub command",
			fields: fields{
				RootCmd: createNestedCommands(2, 2),
			},
			args: args{
				d: createDocument("1 22 "),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "should suggest arguments for registered sub command",
			fields: fields{
				RootCmd: createNestedCommands(2, 2),
			},
			args: args{
				d: createDocument("1 2 "),
			},
			want: []prompt.Suggest{
				newSuggestion("arg"),
			},
		},
		{
			name: "should not suggest for uncompleted command",
			fields: fields{
				RootCmd: createNestedCommands(2, 2),
			},
			args: args{
				d: createDocument("1 2"),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "should not suggest if argument already specified",
			fields: fields{
				RootCmd: createNestedCommands(2, 2),
			},
			args: args{
				d: createDocument("1 2 arg "),
			},
			want: []prompt.Suggest{},
		},
		{
			name: "should suggest after a completed flag",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2)
					cmd.Commands()[0].Flags().StringP("one", "o", "", "Just a flag")
					cmd.Commands()[0].Flags().StringP("two", "t", "", "Just another flag")
					cmd.Commands()[0].Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: createDocument("1 2 --one completed "),
			},
			want: []prompt.Suggest{
				newSuggestion("arg"),
			},
		},
		{
			name: "should not suggest after an uncompleted flag",
			fields: fields{
				RootCmd: func() *cobra.Command {
					cmd := createNestedCommands(2, 2)
					cmd.Commands()[0].Flags().StringP("one", "o", "", "Just a flag")
					cmd.Commands()[0].Flags().StringP("two", "t", "", "Just another flag")
					cmd.Commands()[0].Flags().SortFlags = false
					return cmd
				}(),
			},
			args: args{
				d: createDocument("1 2 --one uncompleted"),
			},
			want: []prompt.Suggest{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewServerSideCompleter(tt.fields.RootCmd)
			addSuggestionFunctions(c, tt.fields.RootCmd)
			c.Complete(createDocument("1 "))   // preload the suggestions
			time.Sleep(100 * time.Millisecond) // let goroutine run
			got := c.Complete(tt.args.d)
			require.Equal(t, tt.want, got)
		})
	}
}

func addSuggestionFunctions(c ServerSideCompleter, rootCmd *cobra.Command) {
	for _, cmd := range rootCmd.Commands() {
		testCmd := &testCommand{
			Command: cmd,
		}
		for i, subCmd := range cmd.Commands() {
			// register only every other sub command
			if i%2 == 0 {
				testCmd.completableChildren = append(testCmd.completableChildren, subCmd)
			}
		}
		c.AddCommand(testCmd)
	}
}

type testCommand struct {
	*cobra.Command
	completableChildren []*cobra.Command
}

func (c *testCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *testCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *testCommand) ServerComplete() []prompt.Suggest {
	return []prompt.Suggest{newSuggestion("arg")}
}
