package output

type Output int

const (
	Human Output = iota
	JSON
	YAML
)

var ValidFlagValues = []string{"human", "json", "yaml"}

func (o Output) String() string {
	return ValidFlagValues[o]
}
