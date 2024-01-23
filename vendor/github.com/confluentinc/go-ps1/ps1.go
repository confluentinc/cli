package ps1

const fmtLongDescription = `Use this command to add ` + "`" + `%s` + "`" + ` information to your terminal prompt.

Bash:

::

	export PS1='$(%s prompt) '$PS1

ZSH:

::

	setopt prompt_subst
	export PS1='$(%s prompt) '$PS1

You can customize the prompt by calling passing the ` + "`" + `--format` + "`" + ` flag, for example ` + "`" + `-f '(%s|%%C)'` + "`" + `.
To make this permanent, you must add the above lines to your Bash or ZSH profile.

Formatting Tokens
~~~~~~~~~~~~~~~~~

This command comes with a number of formatting tokens. What follows is a list of all tokens:

%s

Style
~~~~~

The style of the text can be changed with a combination of functions, colors, and attributes.

Functions:

* fgcolor - Change the foreground color.
* bgcolor - Change the background color.
* attr    - Change a text attribute.

Colors:

* black
* blue
* cyan
* green
* magenta
* red
* white
* yellow

Text Attributes:

* bold
* invert
* italicize
* underline

Examples
~~~~~~~~

* {{fgcolor "blue" "this text is blue"}}
* {{bgcolor "blue" "this text has a blue background"}}
* {{attr "bold" "this text is bold"}}

Use a vertical bar to separate further attributes:

* {{fgcolor "red" "this text is red and bold" | attr "bold"}}

We can use tokens and colors in the same format string:

* ({{fgcolor "blue" "%s"}} | {{fgcolor "red" "%%C"}})`

type ps1 struct {
	cliName string
	tokens  []Token
}

// New builds a PS1 object with a CLI name and custom formatting tokens.
func New(cliName string, tokens []Token) *ps1 {
	if cliName == "" {
		panic("must provide ps1 with a CLI name")
	}

	if tokens == nil || len(tokens) == 0 {
		panic("must provide ps1 with tokens")
	}

	return &ps1{
		cliName: cliName,
		tokens:  tokens,
	}
}
