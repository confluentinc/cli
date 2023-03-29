package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveStatementTerminator(t *testing.T) {
	type args struct {
		statement string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "removeStatementTerminator() removes one terminator", args: args{statement: "SELECT * FROM table;"}, want: "SELECT * FROM table"},
		{name: "removeStatementTerminator() removes no terminator", args: args{statement: "SELECT * FROM table"}, want: "SELECT * FROM table"},
		{name: "removeStatementTerminator() removes multiple terminators", args: args{statement: "SELECT * FROM table;;;"}, want: "SELECT * FROM table"},
		{name: "removeStatementTerminator() doesn't remove terminators in between", args: args{statement: "SELECT * FROM table;;;wasas"}, want: "SELECT * FROM table;;;wasas"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeStatementTerminator(tt.args.statement); got != tt.want {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRemoveWhiteSpaces(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "removeTabNewLineAndWhitesSpaces() removes all whitespaces", args: args{str: " key=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all whitespaces", args: args{str: " key  =    value "}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: " key\n=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\nkey=\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\nvalue\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\nkey=\nvalue\n"}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\r\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: " key\r\n=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\r\nkey=\r\nvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "key=\r\nvalue\r\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all newlines", args: args{str: "\r\nkey=\r\nvalue\r\n"}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "key=\tvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: " key\t=value "}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "\tkey=\tvalue"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "key=\tvalue\t"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all tabs", args: args{str: "\tkey=\tvalue\t"}, want: "key=value"},

		{name: "removeTabNewLineAndWhitesSpaces() removes all new lines, tabs and whitespaces", args: args{str: "\n \tkey\n=\n\tvalue\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all new lines, tabs and whitespaces", args: args{str: "\r\n \tkey\t=\t\tvalue\n"}, want: "key=value"},
		{name: "removeTabNewLineAndWhitesSpaces() removes all new lines, tabs and whitespaces", args: args{str: "\n \tkey\n = \n\tvalue\r\n"}, want: "key=value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeTabNewLineAndWhitesSpaces(tt.args.str); got != tt.want {
				require.Equal(t, tt.want, got)
			}
		})
	}
}
