package ccloud

import (
	"bytes"
	"fmt"
	math "math"
	"runtime"
	"sort"
	"strings"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

var _ error = &Error{}

type Errorer interface {
	GetError() *Error
}

// ReplyErr wraps reply and its error and returns the error if they
// are non-nil.
func ReplyErr(resp Errorer, err error) error {
	if err != nil {
		switch err.(type) {
		case *Error:
			return err
		default:
			return WrapCoreErr(err, "reply error")
		}
	}
	// can't return resp.GetError() cuz of https://golang.org/doc/faq#nil_error
	if err := resp.GetError(); err != nil {
		return err
	}
	return nil
}

// formatURL resolves a url from the provided format string and provided arguments
func formatURL(endpoint string, arguments ...interface{}) string {
	return fmt.Sprintf(string(endpoint), arguments...)
}

// isValidResource checks for invalid resources that are guaranteed not to be found in the backend.
// If a resource is invalid, we don't need to send a request (see CLI-1637).
func isValidResource(resource string) bool {
	return resource != "" && !strings.HasSuffix(resource, "\r")
}

// Error implements the error interface.
func (e *Error) Error() string {
	b := new(bytes.Buffer)
	e.printStack(b)
	pad(b, ": ")
	b.WriteString(e.Message)
	e.writeNested(b)
	if b.Len() == 0 {
		return "no error"
	}
	return b.String()
}

// E is a useful func for instantiating corev1.Errors.
func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to E with no arguments")
	}
	e := &Error{}
	b := new(bytes.Buffer)
	var stack bool
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			pad(b, ": ")
			b.WriteString(arg)
		case int32:
			e.Code = arg
		case int:
			e.Code = int32(arg)
		case bool:
			stack = true
		case error:
			pad(b, ": ")
			b.WriteString(arg.Error())
		}
	}
	e.Message = b.String()
	if stack {
		e.populateStack()
	}
	return e
}

// populateStack uses the runtime to populate the Error's stack struct with
// information about the current stack.
func (e *Error) populateStack() {
	e.Stack = &Stack{Callers: callers()}
}

// pad appends str to the buffer if the buffer already has some data.
func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(str)
}

func (e *Error) writeNested(b *bytes.Buffer) {
	if len(e.NestedErrors) == 0 {
		return
	}
	pad(b, ":")
	var keys []string
	for key := range e.NestedErrors {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, k := range keys {
		pad(b, "\n\t")
		b.WriteString(k)
		pad(b, ": ")
		b.WriteString(e.NestedErrors[k])
	}
}

// frame returns the nth frame, with the frame at top of stack being 0.
func frame(callers []uintptr, n int) *runtime.Frame {
	frames := runtime.CallersFrames(callers)
	var f runtime.Frame
	for i := len(callers) - 1; i >= n; i-- {
		var ok bool
		f, ok = frames.Next()
		if !ok {
			break // Should never happen, and this is just debugging.
		}
	}
	return &f
}

// callers is a wrapper for runtime.callers that allocates a slice.
func callers() []uintptr {
	var stk [64]uintptr
	const skip = 4 // Skip 4 stack frames; ok for both E and Error funcs.
	n := runtime.Callers(skip, stk[:])
	return stk[:n]
}

var separator = ":\n\t"

// printStack formats and prints the stack for this Error to the given buffer.
// It should be called from the Error's Error method.
func (e *Error) printStack(b *bytes.Buffer) {
	if e.Stack == nil {
		return
	}

	printCallers := callers()

	// Iterate backward through e.Stack.Callers (the last in the stack is the
	// earliest call, such as main) skipping over the PCs that are shared
	// by the error stack and by this function call stack, printing the
	// names of the functions and their file names and line numbers.
	var prev string // the name of the last-seen function
	var diff bool   // do the print and error call stacks differ now?
	for i := 0; i < len(e.Stack.Callers); i++ {
		thisFrame := frame(e.Stack.Callers, i)
		name := runtime.FuncForPC(thisFrame.PC).Name()

		if !diff && i < len(printCallers) {
			if name == runtime.FuncForPC(frame(printCallers, i).PC).Name() {
				// both stacks share this PC, skip it.
				continue
			}
			// No match, don't consider printCallers again.
			diff = true
		}

		// Don't print the same function twice.
		// (Can happen when multiple error stacks have been coalesced.)
		if name == prev {
			continue
		}

		// Find the uncommon prefix between this and the previous
		// function name, separating by dots and slashes.
		trim := 0
		for {
			j := strings.IndexAny(name[trim:], "./")
			if j < 0 {
				break
			}
			if !strings.HasPrefix(prev, name[:j+trim]) {
				break
			}
			trim += j + 1 // skip over the separator
		}

		// Do the printing.
		pad(b, separator)
		fmt.Fprintf(b, "%v:%d: ", thisFrame.File, thisFrame.Line)
		if trim > 0 {
			b.WriteString("...")
		}
		b.WriteString(name[trim:])

		prev = name
	}
}

type Stack struct {
	Callers []uintptr
}

func (t *Stack) Size() int {
	return 0
}

func (t *Stack) Unmarshal(data []byte) error {
	return nil
}

func (t *Stack) MarshalTo(data []byte) (n int, err error) {
	return 0, nil
}
