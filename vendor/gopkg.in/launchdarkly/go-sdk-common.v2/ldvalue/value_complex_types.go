package ldvalue

// This file contains types and methods that are only used for complex data structures (array and
// object), in the fully immutable model where no slices, maps, or interface{} values are exposed.

// ArrayBuilder is a builder created by ArrayBuild(), for creating immutable arrays.
//
// An ArrayBuilder should not be accessed by multiple goroutines at once.
type ArrayBuilder interface {
	// Add appends an element to the array builder.
	Add(value Value) ArrayBuilder
	// Build creates a Value containing the previously added array elements. Continuing to modify the
	// same builder by calling Add after that point does not affect the returned array.
	Build() Value
}

type arrayBuilderImpl struct {
	builder valueArrayBuilderImpl
}

// ObjectBuilder is a builder created by ObjectBuild(), for creating immutable JSON objects.
//
// An ObjectBuilder should not be accessed by multiple goroutines at once.
type ObjectBuilder interface {
	// Set sets a key-value pair in the object builder.
	Set(key string, value Value) ObjectBuilder
	// Build creates a Value containing the previously specified key-value pairs. Continuing to modify
	// the same builder by calling Set after that point does not affect the returned object.
	Build() Value
}

type objectBuilderImpl struct {
	builder valueMapBuilderImpl
}

// ArrayOf creates an array Value from a list of Values.
//
// This requires a slice copy to ensure immutability; otherwise, an existing slice could be passed
// using the spread operator, and then modified. However, since Value is itself immutable, it does
// not need to deep-copy each item.
func ArrayOf(items ...Value) Value {
	return Value{valueType: ArrayType, arrayValue: ValueArrayOf(items...)}
}

// ArrayBuild creates a builder for constructing an immutable array Value.
//
//     arrayValue := ldvalue.ArrayBuild().Add(ldvalue.Int(100)).Add(ldvalue.Int(200)).Build()
func ArrayBuild() ArrayBuilder {
	return ArrayBuildWithCapacity(1)
}

// ArrayBuildWithCapacity creates a builder for constructing an immutable array Value.
//
// The capacity parameter is the same as the capacity of a slice, allowing you to preallocate space
// if you know the number of elements; otherwise you can pass zero.
//
//     arrayValue := ldvalue.ArrayBuildWithCapacity(2).Add(ldvalue.Int(100)).Add(ldvalue.Int(200)).Build()
func ArrayBuildWithCapacity(capacity int) ArrayBuilder {
	return &arrayBuilderImpl{valueArrayBuilderImpl{output: make([]Value, 0, capacity)}}
}

func (b *arrayBuilderImpl) Add(value Value) ArrayBuilder {
	b.builder.Add(value)
	return b
}

func (b *arrayBuilderImpl) Build() Value {
	return Value{valueType: ArrayType, arrayValue: b.builder.Build()}
}

// CopyObject creates a Value by copying an existing map[string]Value.
//
// If you want to copy a map[string]interface{} instead, use CopyArbitraryValue.
func CopyObject(m map[string]Value) Value {
	return Value{valueType: ObjectType, objectValue: CopyValueMap(m)}
}

// ObjectBuild creates a builder for constructing an immutable JSON object Value.
//
//     objValue := ldvalue.ObjectBuild().Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ObjectBuild() ObjectBuilder {
	return ObjectBuildWithCapacity(1)
}

// ObjectBuildWithCapacity creates a builder for constructing an immutable JSON object Value.
//
// The capacity parameter is the same as the capacity of a map, allowing you to preallocate space
// if you know the number of elements; otherwise you can pass zero.
//
//     objValue := ldvalue.ObjectBuildWithCapacity(2).Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ObjectBuildWithCapacity(capacity int) ObjectBuilder {
	return &objectBuilderImpl{valueMapBuilderImpl{output: make(map[string]Value, capacity)}}
}

func (b *objectBuilderImpl) Set(name string, value Value) ObjectBuilder {
	b.builder.Set(name, value)
	return b
}

func (b *objectBuilderImpl) Build() Value {
	return Value{valueType: ObjectType, objectValue: b.builder.Build()}
}

// Count returns the number of elements in an array or JSON object.
//
// For values of any other type, it returns zero.
func (v Value) Count() int {
	switch v.valueType {
	case ArrayType:
		return v.arrayValue.Count()
	case ObjectType:
		return v.objectValue.Count()
	}
	return 0
}

// GetByIndex gets an element of an array by index.
//
// If the value is not an array, or if the index is out of range, it returns Null().
func (v Value) GetByIndex(index int) Value {
	ret, _ := v.TryGetByIndex(index)
	return ret
}

// TryGetByIndex gets an element of an array by index, with a second return value of true if
// successful.
//
// If the value is not an array, or if the index is out of range, it returns (Null(), false).
func (v Value) TryGetByIndex(index int) (Value, bool) {
	return v.arrayValue.TryGet(index)
	// This is always safe because if v isn't an array, arrayValue is an empty ValueArray{}
	// and TryGet will always return Null(), false.
}

// Keys returns the keys of a JSON object as a slice.
//
// The method copies the keys. If the value is not an object, it returns an empty slice.
func (v Value) Keys() []string {
	if v.valueType == ObjectType {
		return v.objectValue.Keys()
	}
	return nil
}

// GetByKey gets a value from a JSON object by key.
//
// If the value is not an object, or if the key is not found, it returns Null().
func (v Value) GetByKey(name string) Value {
	return v.objectValue.Get(name)
	// This is always safe because if v isn't an object, objectValue is an empty ValueMap{}
	// and keys will never be found.
}

// TryGetByKey gets a value from a JSON object by key, with a second return value of true if
// successful.
//
// If the value is not an object, or if the key is not found, it returns (Null(), false).
func (v Value) TryGetByKey(name string) (Value, bool) {
	return v.objectValue.TryGet(name)
}

// Enumerate calls a function for each value within a Value.
//
// If the input value is Null(), the function is not called.
//
// If the input value is an array, fn is called for each element, with the element's index in the
// first parameter, "" in the second, and the element value in the third.
//
// If the input value is an object, fn is called for each key-value pair, with 0 in the first
// parameter, the key in the second, and the value in the third.
//
// For any other value type, fn is called once for that value.
//
// The return value of fn is true to continue enumerating, false to stop.
func (v Value) Enumerate(fn func(index int, key string, value Value) bool) {
	switch v.valueType {
	case NullType:
		return
	case ArrayType:
		for i, v1 := range v.arrayValue.data {
			if !fn(i, "", v1) {
				return
			}
		}
	case ObjectType:
		for k, v1 := range v.objectValue.data {
			if !fn(0, k, v1) {
				return
			}
		}
	default:
		_ = fn(0, "", v)
	}
}

// Transform applies a transformation function to a Value, returning a new Value.
//
// The behavior is as follows:
//
// If the input value is Null(), the return value is always Null() and the function is not called.
//
// If the input value is an array, fn is called for each element, with the element's index in the
// first parameter, "" in the second, and the element value in the third. The return values of fn
// can be either a transformed value and true, or any value and false to remove the element.
//
//     ldvalue.ArrayOf(ldvalue.Int(1), ldvalue.Int(2), ldvalue.Int(3)).Build().
//         Transform(func(index int, key string, value Value) (Value, bool) {
//             if value.IntValue() == 2 {
//                 return ldvalue.Null(), false
//             }
//             return ldvalue.Int(value.IntValue() * 10), true
//         })
//     // returns [10, 30]
//
// If the input value is an object, fn is called for each key-value pair, with 0 in the first
// parameter, the key in the second, and the value in the third. Again, fn can choose to either
// transform or drop the value.
//
//     ldvalue.ObjectBuild().Set("a", ldvalue.Int(1)).Set("b", ldvalue.Int(2)).Set("c", ldvalue.Int(3)).Build().
//         Transform(func(index int, key string, value Value) (Value, bool) {
//             if key == "b" {
//                 return ldvalue.Null(), false
//             }
//             return ldvalue.Int(value.IntValue() * 10), true
//         })
//     // returns {"a": 10, "c": 30}
//
// For any other value type, fn is called once for that value; if it provides a transformed
// value and true, the transformed value is returned, otherwise Null().
//
//     ldvalue.Int(2).Transform(func(index int, key string, value Value) (Value, bool) {
//         return ldvalue.Int(value.IntValue() * 10), true
//     })
//     // returns numeric value of 20
//
// For array and object values, if the function does not modify or drop any values, the exact
// same instance is returned without allocating a new slice or map.
func (v Value) Transform(fn func(index int, key string, value Value) (Value, bool)) Value {
	switch v.valueType {
	case NullType:
		return v
	case ArrayType:
		return Value{valueType: ArrayType, arrayValue: v.arrayValue.Transform(
			func(index int, value Value) (Value, bool) {
				return fn(index, "", value)
			},
		)}
	case ObjectType:
		return Value{valueType: ObjectType, objectValue: v.objectValue.Transform(
			func(key string, value Value) (string, Value, bool) {
				resultValue, ok := fn(0, key, value)
				return key, resultValue, ok
			},
		)}
	default:
		if transformedValue, ok := fn(0, "", v); ok {
			return transformedValue
		}
		return Null()
	}
}

// AsValueArray converts the Value to the immutable ValueArray type if it is a JSON array.
// Otherwise it returns a ValueArray in an uninitialized (nil) state. This is an efficient operation
// that does not allocate a new slice.
func (v Value) AsValueArray() ValueArray {
	return v.arrayValue
}

// AsValueMap converts the Value to the immutable ValueMap type if it is a JSON object. Otherwise
// it returns a ValueMap in an uninitialized (nil) state. This is an efficient operation that does
// not allocate a new map.
func (v Value) AsValueMap() ValueMap {
	return v.objectValue
}
