package mock

func NewPromptMock(lines ...string) *Prompt {
	i := 0

	return &Prompt{
		ReadLineFunc: func() (string, error) {
			line := lines[i]
			i++
			return line, nil
		},
		ReadLineMaskedFunc: func() (string, error) {
			line := lines[i]
			i++
			return line, nil
		},
	}
}
