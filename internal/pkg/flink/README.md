### flink-sql-client

Flink SQL Client to be used with Confluent Cloud. Powered by go-prompt and tview.

#### Run prototype with static mock

Install dependencies

```
make deps
```

Run prototype

```
go run _examples/main/demo_main.go
```

#### Local properties

We'll add a couple of local properties to configure the client which aren't flink related and will only exist in the client. We'll for now document these here until we have a official documentation for the client.

| Property | Description | Default |
| table.results-timeout | the total amount of time in seconds to wait before timing out the request waiting for results to be ready | 780 (13 min) |


#### Building for other operation systems

You can build the `./demo_main` executable with the command `go build ./_examples/devel/demo_main.go`.

If you want to build for other operating systems and architectures, set the `GOOS` and `GOARCH` environment variables

e.g. to build for windows

```sh
GOOS=windows GOARCH=amd64 go build ./_examples/devel/demo_main.go
```

You won't be able to run the binary unless you have a windows machine/VM, but at least you can test if the binary builds on windows.
