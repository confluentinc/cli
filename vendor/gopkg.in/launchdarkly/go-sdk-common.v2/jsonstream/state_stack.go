package jsonstream

type streamState int

const (
	stateValueFirst     = iota // no top-level values have been written
	stateValueNext      = iota // at least one top-level value has been written
	stateArrayStart     = iota // array has been started, has no values yet
	stateArrayNext      = iota // array has been started and has at least one value
	stateObjectStart    = iota // object has been started, has no names or values yet
	stateObjectNameNext = iota // object has been started, has name-value pairs, can write a new name
	stateObjectValue    = iota // object has been started, name has been written, needs a value
)

type stateStack struct {
	current  streamState
	array    [20]streamState // we use this fixed-size array whenever possible to avoid heap allocations
	arrayLen int
	slice    []streamState // this slice gets allocated on the heap only if necessary
}

func (s *stateStack) push(newState streamState) {
	// The separate logic here for the slice and the array allows us to avoid allocating a backing array for a
	// slice on the heap unless we run out of room in our local array. The original implementation of this tried
	// to be clever by initializing the slice to refer to stateStackArray[0:0]; that would work if we were only
	// using the slice within the same scope where it was initialized or a deeper scope, but since it stays
	// around after this method returns, the compiler would consider it suspicious enough to cause escaping.
	if s.slice == nil {
		if s.arrayLen < len(s.array) {
			s.array[s.arrayLen] = s.current
			s.arrayLen++
		} else {
			s.slice = make([]streamState, s.arrayLen, s.arrayLen*2)
			copy(s.slice, s.array[:])
			s.slice = append(s.slice, s.current)
		}
	} else {
		s.slice = append(s.slice, s.current)
	}
	s.current = newState
}

func (s *stateStack) pop() {
	if s.slice != nil {
		n := len(s.slice)
		s.current = s.slice[n-1]
		s.slice = s.slice[0 : n-1]
	} else {
		s.current = s.array[s.arrayLen-1]
		s.arrayLen--
	}
}

func (s *stateStack) isTopLevel() bool {
	if s.slice == nil {
		return s.arrayLen == 0
	}
	return len(s.slice) == 0
}
