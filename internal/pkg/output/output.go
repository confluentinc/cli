package output

const (
	Human output = iota
	JSON
	YAML
)

const FlagName = "output"

var ValidFlagValues = []string{"human", "json", "yaml"}

type output int

func (o output) String() string {
	return ValidFlagValues[o]
}
