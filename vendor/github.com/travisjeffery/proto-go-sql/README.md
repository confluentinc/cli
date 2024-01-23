# proto-go-sql

Generate sql.Scanner and driver.Valuer implementations for your Protobufs.

## Example

We want the generated struct for this Person message to implement sql.Scanner and driver.Valuer so we can easily write and read it as JSON from Postgres.

So we compile the person.proto file:

``` proto
syntax = "proto3";

import "github.com/travisjeffery/proto-go-sql/sql.proto";

message Person {
  option (sql.all) = "json";

  string id = 1;
}
```

And run:

``` sh
$ protoc --sql_out=. person.proto
```

Generating this person_sql.go:

``` go
func (t *Person) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), t)
}

func (t *Person) Value() (driver.Value, error) {
	return json.Marshal(t)
}

```

And we're done!

## License

MIT

---

- [travisjeffery.com](http://travisjeffery.com)
- GitHub [@travisjeffery](https://github.com/travisjeffery)
- Twitter [@travisjeffery](https://twitter.com/travisjeffery)
- Medium [@travisjeffery](https://medium.com/@travisjeffery)
