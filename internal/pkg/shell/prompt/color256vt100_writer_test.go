package prompt

import (
	"bytes"
	"reflect"
	"testing"

	gopropmt "github.com/c-bata/go-prompt"
)

func Test_writeColorString(t *testing.T) {
	type args struct {
		fg    gopropmt.Color
		bg    gopropmt.Color
		attrs []gopropmt.DisplayAttribute
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{
			name: "write color string with default foreground color",
			args: args{
				fg:    0,
				bg:    123,
				attrs: nil,
			},
			wantW: "\x1b[39;48;5;123m",
		},
		{
			name: "write color string with default background color",
			args: args{
				fg:    123,
				bg:    0,
				attrs: nil,
			},
			wantW: "\x1b[38;5;123;49m",
		},
		{
			name: "write color string with no default colors",
			args: args{
				fg:    123,
				bg:    456,
				attrs: nil,
			},
			wantW: "\x1b[38;5;123;48;5;456m",
		},
		{
			name: "write color string with display attributes",
			args: args{
				fg:    123,
				bg:    456,
				attrs: []gopropmt.DisplayAttribute{89},
			},
			wantW: "\x1b[89;38;5;123;48;5;456m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			writeColorString(w, tt.args.fg, tt.args.bg, tt.args.attrs...)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("writeColorString() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func Test_displayAttributeToBytes(t *testing.T) {
	type args struct {
		attribute gopropmt.DisplayAttribute
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "succeed converting valid DisplayAttribute",
			args: args{attribute: 89},
			want: []byte{'8', '9'},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayAttributeToBytes(tt.args.attribute); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("displayAttributeToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
