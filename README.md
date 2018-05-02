# Confluent CLI

## Developing

```
$ make compile-proto
$ go run main.go --help
```

The CLI automatically adds commands when their respective plugins are installed. Enabling the connect
commands by installing the plugins:

```
$ make install-plugins
```

Now you can run:

```
$ go run main.go connect list
```
