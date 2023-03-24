### flink-sql-client

Flink SQL Client to be used with Confluent Cloud. Powered by go-prompt and tview.

**Experimental**: There are still elements around this client yet to be finished like the design and the gateway with whom it will communicate. So we're starting with a lot of moving parts trying to keep moving while waiting on those. The code is experimental and will go through a couple of refactorings.

#### Run prototype

We are using go version 1.19 in this repo.

Install dependencies

```
make deps
```

Run prototype

```
go run _examples/demo_main.go
```

#### Contributing

Take a look at our [CONTRIBUTING.md](./CONTRIBUTING.md) for some notes about contributions.
