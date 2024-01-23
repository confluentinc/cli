package jsonstream

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"unicode/utf8"
)

var (
	tokenNull  = []byte("null")  //nolint:gochecknoglobals
	tokenTrue  = []byte("true")  //nolint:gochecknoglobals
	tokenFalse = []byte("false") //nolint:gochecknoglobals
)

// JSONBuffer is a fast JSON encoder for manual generation of sequential output, writing one token at a
// time. Output is written to an in-memory buffer.
//
// Any invalid operation (such as trying to write a property name when a value is expected, or vice versa)
// causes the JSONBuffer to enter a failed state where all subsequent write operations are ignored. The
// error will be returned by Get() or GetError().
//
// If the caller write smore than one JSON value at the top level (that is, not inside an array or object),
// the values will be separated by whatever byte sequence was specified with SetSeparator(), or, if not
// specified, by a newline.
//
// JSONBuffer is not safe for concurrent access by multiple goroutines.
//
//     var buf jsonstream.JSONBuffer
//     buf.BeginObject()
//     buf.WriteName("a")
//     buf.WriteInt(2)
//     buf.EndObject()
//     bytes, err := buf.Get() // bytes == []byte(`{"a":2}`)
type JSONBuffer struct {
	buf       streamableBuffer
	state     stateStack
	separator []byte
	err       error
}

// NewJSONBuffer creates a new JSONBuffer on the heap. This is not strictly necessary; declaring a local
// value of JSONBuffer{} will also work.
func NewJSONBuffer() *JSONBuffer {
	return &JSONBuffer{}
}

// NewStreamingJSONBuffer creates a JSONBuffer that, instead of accumulating all of the output in memory,
// writes it in chunks to the specified Writer.
//
// In this mode, operations that write data to the JSONBuffer will accumulate the output in memory until
// either at least chunkSize bytes have been written, or Flush() is called. At that point, the buffered
// output is written to the Writer, and then the buffer is cleared. The amount of data written at a time
// may be more than chunkSize bytes, but will not be less unless you force a Flush().
//
// If the Writer returns an error at any point, the JSONBuffer enters a failed state and will not try to
// write any more data. The error can be checked by calling GetError().
//
// It is important to call Flush() after you are done with the JSONBuffer to ensure that everything has
// been written to the Writer.
func NewStreamingJSONBuffer(w io.Writer, chunkSize int) *JSONBuffer {
	j := &JSONBuffer{}
	j.buf.Grow(chunkSize)
	j.buf.SetStreamingWriter(w, chunkSize)
	return j
}

// Get returns the full encoded byte slice.
//
// If the buffer is in a failed state from a previous invalid operation, or cannot be ended at this point
// because of an unfinished array or object, Get() returns a nil slice and the error. In that case, the
// data written so far can be accessed with GetPartial().
func (j *JSONBuffer) Get() ([]byte, error) {
	if j.err != nil {
		return nil, j.err
	}
	if !j.state.isTopLevel() {
		j.err = errors.New("array or object not ended")
		return nil, j.err
	}
	return j.buf.Bytes(), nil
}

// GetError returns an error if the buffer is in a failed state from a previous invalid operation, or
// nil otherwise.
func (j *JSONBuffer) GetError() error {
	if j.err != nil {
		return j.err
	}
	return j.buf.GetWriterError()
}

// GetPartial returns the data written to the buffer so far, regardless of whether it is in a failed or
// incomplete state.
func (j *JSONBuffer) GetPartial() []byte {
	return j.buf.Bytes()
}

// Grow expands the internal buffer by the specified number of bytes. It is the same as calling Grow
// on a bytes.Buffer.
func (j *JSONBuffer) Grow(n int) {
	j.buf.Grow(n)
}

// Flush writes any remaining in-memory output to the underlying Writer, if this is a streaming buffer
// created with NewStreamingJSONBuffer. It has no effect otherwise.
func (j *JSONBuffer) Flush() error {
	return j.buf.Flush()
}

// SetSeparator specifies a byte sequence that should be added to the buffer in between values if more
// than one value is written outside of an array or object. If not specified, a newline is used.
//
//     var buf jsonstream.JSONBuffer
//     buf.SetSeparator([]byte("! "))
//     buf.WriteInt(1)
//     buf.WriteInt(2)
//     buf.WriteInt(3)
//     bytes, err := buf.Get() // bytes == []byte(`1! 2! 3`)
func (j *JSONBuffer) SetSeparator(separator []byte) {
	if separator == nil {
		j.separator = nil
	} else {
		j.separator = make([]byte, len(separator))
		copy(j.separator, separator)
	}
}

// WriteNull writes a JSON null value to the output.
func (j *JSONBuffer) WriteNull() {
	if !j.beforeValue() {
		return
	}
	j.buf.Write(tokenNull)
	j.afterValue()
}

// WriteBool writes a JSON boolean value to the output.
func (j *JSONBuffer) WriteBool(value bool) {
	if !j.beforeValue() {
		return
	}
	if value {
		j.buf.Write(tokenTrue)
	} else {
		j.buf.Write(tokenFalse)
	}
	j.afterValue()
}

// WriteInt writes a JSON numeric value to the output.
func (j *JSONBuffer) WriteInt(value int) {
	if !j.beforeValue() {
		return
	}

	if value == 0 {
		j.buf.WriteRune('0')
	} else {
		byteSlice := make([]byte, 0, 11) // preallocate on stack with room for any numeric string of this size
		byteSlice = strconv.AppendInt(byteSlice, int64(value), 10)
		j.buf.Write(byteSlice)
	}

	j.afterValue()
}

// WriteUint64 writes a JSON numeric value to the output.
func (j *JSONBuffer) WriteUint64(value uint64) {
	if !j.beforeValue() {
		return
	}

	if value == 0 {
		j.buf.WriteRune('0')
	} else {
		byteSlice := make([]byte, 0, 25) // preallocate on stack with room for any numeric string of this size
		byteSlice = strconv.AppendUint(byteSlice, value, 10)
		j.buf.Write(byteSlice)
	}

	j.afterValue()
}

// WriteFloat64 writes a JSON numeric value to the output.
func (j *JSONBuffer) WriteFloat64(value float64) {
	if !j.beforeValue() {
		return
	}

	if value == 0 {
		j.buf.WriteRune('0')
	} else {
		byteSlice := make([]byte, 0, 30) // preallocate on stack with room for most numeric strings of this size
		// (due to how append works, if it happens *not* to be big enough, byteSlice will just escape to the heap)

		byteSlice = strconv.AppendFloat(byteSlice, value, 'f', -1, 64)
		j.buf.Write(byteSlice)
	}

	j.afterValue()
}

// WriteString writes a JSON string value to the output, with quotes and escaping.
//
// JSONBuffer assumes that multi-byte UTF8 characters are allowed in the output, so it will not escape any
// characters other than control characters, double quotes, and backslashes.
func (j *JSONBuffer) WriteString(value string) {
	if !j.beforeValue() {
		return
	}
	j.writeQuotedString(value)
	j.afterValue()
}

// WriteRaw writes a pre-encoded JSON value to the output as-is. Its format is assumed to be correct;
// this operation will not fail unless it is not permitted to write a value at this point.
func (j *JSONBuffer) WriteRaw(value json.RawMessage) {
	if !j.beforeValue() {
		return
	}
	j.buf.Write(value)
	j.afterValue()
}

// BeginArray begins writing a JSON array.
//
// All subsequent values written will be delimited by commas. Call EndArray to finish the array. The
// array may contain any types of values, including nested arrays or objects.
//
//     buf.BeginArray()
//     buf.WriteInt(1)
//     buf.WriteString("b")
//     buf.EndArray() // produces [1,"b"]
func (j *JSONBuffer) BeginArray() {
	if !j.beforeValue() {
		return
	}
	j.buf.WriteRune('[')
	j.state.push(stateArrayStart)
}

// EndArray finishes writing the current JSON array.
func (j *JSONBuffer) EndArray() {
	if j.err != nil {
		return
	}
	if j.state.current != stateArrayStart && j.state.current != stateArrayNext {
		j.fail("called EndArray when not inside an array")
		return
	}
	j.buf.WriteRune(']')
	j.state.pop()
	j.afterValue()
}

// BeginObject begins writing a JSON object.
//
// Until this object is ended, you must call WriteName before each value. Call EndObject to finish
// the object. The object may contain any types of values, including nested objects or arrays.
//
//     buf.BeginObject()
//     buf.WriteName("a")
//     buf.WriteInt(1)
//     buf.WriteName("b")
//     buf.WriteBool(true)
//     buf.EndObject() // produces {"a":1,"b":true}
func (j *JSONBuffer) BeginObject() {
	if !j.beforeValue() {
		return
	}
	j.buf.WriteRune('{')
	j.state.push(stateObjectStart)
}

// WriteName writes a property name within an object.
//
// It is an error to call this method outside of an object, or immediately after another WriteName.
// Each WriteName should be followed by some JSON value (WriteBool, WriteString, BeginArray, etc.).
func (j *JSONBuffer) WriteName(name string) {
	if j.err != nil {
		return
	}
	if j.state.current != stateObjectStart && j.state.current != stateObjectNameNext {
		j.fail("called WriteName when a value was expected")
		return
	}
	if j.state.current == stateObjectNameNext {
		j.buf.WriteRune(',')
	}
	j.writeQuotedString(name)
	j.buf.WriteRune(':')
	j.state.current = stateObjectValue
}

// EndObject finishes writing the current JSON object.
func (j *JSONBuffer) EndObject() {
	if j.err != nil {
		return
	}
	if j.state.current == stateObjectValue {
		j.fail("called EndObject when a value was expected")
		return
	}
	if j.state.current != stateObjectStart && j.state.current != stateObjectNameNext {
		j.fail("called EndObject when not inside an object")
		return
	}
	j.buf.WriteRune('}')
	j.state.pop()
	j.afterValue()
}

func (j *JSONBuffer) beforeValue() bool {
	if j.err != nil {
		return false
	}
	switch j.state.current {
	case stateValueNext:
		if j.separator == nil {
			j.buf.WriteByte('\n')
		} else {
			j.buf.Write(j.separator)
		}
	case stateArrayNext:
		j.buf.WriteByte(',')
	case stateObjectStart:
		j.fail("wrote value when property name was expected")
		return false
	case stateObjectNameNext:
		j.fail("wrote value when property name was expected")
		return false
	}
	return true
}

func (j *JSONBuffer) afterValue() {
	switch j.state.current {
	case stateValueFirst:
		j.state.current = stateValueNext
	case stateArrayStart:
		j.state.current = stateArrayNext
	case stateObjectValue:
		j.state.current = stateObjectNameNext
	}
}

func (j *JSONBuffer) writeQuotedString(s string) {
	// This is basically the same logic used internally by json.Marshal
	j.buf.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		aByte := s[i]
		if aByte < ' ' || aByte == '"' || aByte == '\\' {
			if i > start {
				j.buf.WriteString(s[start:i])
			}
			j.writeEscapedChar(aByte)
			i++
			start = i
		} else {
			if aByte < utf8.RuneSelf { // single-byte character
				i++
			} else {
				_, size := utf8.DecodeRuneInString(s[i:])
				i += size
			}
		}
	}
	if start < len(s) {
		j.buf.WriteString(s[start:])
	}
	j.buf.WriteByte('"')
}

func (j *JSONBuffer) writeEscapedChar(ch byte) {
	j.buf.WriteByte('\\')
	switch ch {
	case '\b':
		j.buf.WriteByte('b')
	case '\t':
		j.buf.WriteByte('t')
	case '\n':
		j.buf.WriteByte('n')
	case '\f':
		j.buf.WriteByte('f')
	case '\r':
		j.buf.WriteByte('r')
	case '"':
		j.buf.WriteByte('"')
	case '\\':
		j.buf.WriteByte('\\')
	default:
		j.buf.WriteString("u00")
		hexChars := make([]byte, 0, 4)
		hexChars = strconv.AppendInt(hexChars, int64(ch), 16)
		if len(hexChars) < 2 {
			j.buf.WriteByte('0')
		}
		j.buf.Write(hexChars)
	}
}

func (j *JSONBuffer) fail(message string) {
	j.err = errors.New(message)
}
