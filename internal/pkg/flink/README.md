### flink-sql-client

Flink SQL Client to be used with Confluent Cloud. Powered by go-prompt and tview.

**Experimental**: There are still elements around this client yet to be finished like the design and the gateway with whom it will communicate. So we're starting with a lot of moving parts trying to keep moving while waiting on those. The code is experimental and will go through a couple of refactorings.

#### Run prototype

We are using go version 1.19 in this repo.

Install dependencies

```
make deps
make update-submodules // for gateway-api submodule
```

Run prototype

```
go run .
```

#### Contributing

Take a look at our [CONTRIBUTING.md](./CONTRIBUTING.md) for some notes about contributions.

## Local Flink sdk

We have a local version of flink minispec and sdk (to iterate faster). If you want to make a change:

```bash
make update-submodules

make update-local-sdk

make deps
```

If you want to change the branch from the submodule, you might need to run:

```
git submodule update --init --recursive --remote
```

The generated sdk will have formatting issue and cause tests to fail (we run go vet), to fix it run the following (you might have to run it multiple times, if you continue to have formatting errors)

```
make fmt
```

If you have python dependencies errors, run:

```
cd gateway_api/minispec
make python-deps

cd gateway_api/aggregator
make python-deps
make node-deps
```

If you have a vendor folder delete it, otherwise the replace won't work
