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
go run .
```

#### Adding custom shortcuts to the client

An important part of a good command line tool are shortcuts. To add custom shortcuts with go-prompt, you can take a look here at:

https://stackoverflow.com/questions/6205157/how-to-set-keyboard-shortcuts-to-jump-to-beginning-end-of-line

As an example, I've added [option + arrow navigation](./main.go#L29) in the current example.

#### Inspirations for testing tview - we prob won't test this though

https://github.com/rivo/tview/pull/441
https://github.com/rivo/tview/pull/5/
