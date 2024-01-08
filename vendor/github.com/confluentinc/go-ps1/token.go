package ps1

import "fmt"

// Token represents a formatting token: a single-character symbol that maps to a generic function that returns a string.
type Token struct {
	Name byte
	Desc string
	Func func() string
}

func (t Token) String() string {
	return fmt.Sprintf("%%%c", t.Name)
}
