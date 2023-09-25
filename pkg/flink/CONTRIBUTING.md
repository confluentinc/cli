### Run tests

You can either run tests for the whole app by running `make deps` from the root directory or `go test ./...` from a specific package.

### Run test converage

The following command runs the tests and generates a coverage report in the `coverage` folder.
```
go test -coverprofile=coverage.out ./... && \
go tool cover -html=coverage.out -o coverage.html
```
You can then open the coverage.html file from the repository root in your browser to see the coverage report for each file.

### Generating Mocks
We generate mocks for interfaces as we need them. If you update an interface, you might have to generate mocks again. Make sure you have mockgen installed:

```
go install go.uber.org/mock/mockgen@latest
```

Generate mocks:
```
go generate pkg/flink/test/mock/mock_generator.go
```

If you need to generate mocks for a new interface, take a look at the mock_generator.go and change it accordingly.

More info on how to generate mocks: https://github.com/uber-go/mock

#### Adding custom shortcuts to the client

An important part of a good command line tool are shortcuts. To add custom shortcuts with go-prompt, you can take a look here at:

https://stackoverflow.com/questions/6205157/how-to-set-keyboard-shortcuts-to-jump-to-beginning-end-of-line

We have a couple of examples in the codebase.

### We have Snapshot tests for some components

You can learn more about it here: https://github.com/bradleyjkemp/cupaloy

When failing, you might need to review your snapshots and update them if they look good. You can easily do that by running:

```
UPDATE_SNAPSHOTS=true go test ./...
```
