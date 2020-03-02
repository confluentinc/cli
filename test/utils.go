package test

import "bytes"

// NormalizeNewLines replaces \r\n and \r newline sequences with \n
func NormalizeNewLines(raw string) string {
	normalized := bytes.Replace([]byte(raw), []byte{13, 10}, []byte{10}, -1)
	normalized = bytes.Replace(normalized, []byte{13}, []byte{10}, -1)
	return string(normalized)
}
