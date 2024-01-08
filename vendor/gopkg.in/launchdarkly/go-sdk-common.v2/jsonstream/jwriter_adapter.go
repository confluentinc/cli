package jsonstream

import (
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

type bufferAdapter struct {
	jsonBuffer *JSONBuffer
}

func (b *bufferAdapter) Write(data []byte) (int, error) {
	b.jsonBuffer.buf.Write(data)
	if err := b.jsonBuffer.GetError(); err != nil {
		return 0, err // COVERAGE: can't reproduce this condition in unit tests
	}
	return len(data), nil
}

// WriteToJSONBufferThroughWriter is a convenience method that allows marshaling logic written against
// the newer jsonstream API to be used with this deprecated package.
func WriteToJSONBufferThroughWriter(writable jwriter.Writable, jsonBuffer *JSONBuffer) {
	jsonBuffer.beforeValue()
	b := bufferAdapter{jsonBuffer}
	w := jwriter.NewStreamingWriter(&b, 100)
	writable.WriteToJSONWriter(&w)
	_ = w.Flush()
	err := w.Error()
	if err != nil && jsonBuffer.err == nil {
		jsonBuffer.err = err
	}
	jsonBuffer.afterValue()
}
