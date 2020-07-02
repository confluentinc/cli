package errors

import "bytes"

func GetErrorStringWithSuggestions(err error) string {
	var b bytes.Buffer
	DisplaySuggestionsMessage(err, &b)
	out := b.String()
	if out == "" {
		return err.Error()
	}
	return err.Error() + "\n" + out
}
