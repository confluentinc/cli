### flink

This package contains Flink SQL Client to be used with Confluent Cloud and other sources that will be used for interacting with the Flink Gateway. Powered by go-prompt and tview.

#### Run prototype with static mock

Install dependencies

```
go mod download
go mod verify
```

Run prototype

```
go run _examples/main/demo_main.go
```

#### Local properties

We'll add a couple of local properties to configure the client which aren't flink related and will only exist in the client. We'll for now document these here until we have a official documentation for the client.

| Property | Description | Default |
| table.results-timeout | the total amount of time in seconds to wait before timing out the request waiting for results to be ready | 600 (10 min) |

