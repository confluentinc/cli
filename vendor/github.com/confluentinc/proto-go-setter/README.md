# proto-go-setter

Add setters for generated message fields. Supports scalar values (string, int32, etc) and maps
with primitive types for both key and value.

This is an often-requested feature for protoc-gen-go that won't be added to the core generator. See:

* https://github.com/golang/protobuf/issues/65
* https://github.com/golang/protobuf/issues/664

## Example

The simplest option is to just generate a setter for a single field.

```proto
syntax = "proto3";

import "github.com/confluentinc/proto-go-setter/setter.proto";

message Person {
  string id = 1;
  string name = 2 [(setter.include)=true];
}
```

Alternatively, you may want to generate setters for all fields in a message, or
all fields _except_ a single field.

```proto
syntax = "proto3";

import "github.com/confluentinc/proto-go-setter/setter.proto";

message Person {
  option (setter.all_fields) = true;

  string id = 1 [(setter.exclude)=true];
  string name = 2;
}
```

Lastly, you may want to generate setters for everything in a file.

```proto
syntax = "proto3";

import "github.com/confluentinc/proto-go-setter/setter.proto";

option (setter.all_messages) = true;

message Person {
  string id = 1 [(setter.exclude)=true];
  string name = 2;
}
```

You'd generate the setters code by running

```bash
$ protoc --setter_out=. person.proto
```

All three examples above would result in `person_setter.go` containing

```go
func (t *Person) SetName(name string) {
	t.Name = name
}
```

## LICENSE

MIT

---

- [codyaray.com](http://codyaray.com)
- GitHub [@codyaray](https://github.com/codyaray)
- LinkedIn [@codyaray](https://linkedin.com/in/codyaray)
