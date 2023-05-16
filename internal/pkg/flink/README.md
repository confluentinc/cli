### flink-sql-client

Flink SQL Client to be used with Confluent Cloud. Powered by go-prompt and tview.

**Experimental**: There are still elements around this client yet to be finished like the design and the gateway with whom it will communicate. So we're starting with a lot of moving parts trying to keep moving while waiting on those. The code is experimental and will go through a couple of refactorings.

#### Go version
We are currently using go version 1.20 in this repo.

#### Run prototype with static mock

Install dependencies

```
make deps
```

Run prototype

```
go run _examples/main/demo_main.go
```

#### Run demo in devel

We have can run our client against a deployed gateway service running in our devel environment. You can find more information about this [on slide 11](https://docs.google.com/presentation/d/1EARZ8hXm9i5h9p2OnjDVRMWWdEXyOMOfWZ0tcbF6tJo/edit#slide=id.g227e6404467_0_156). To run the demo, simply follow the steps below:

````
make deps
go run _examples/devel/demo_devel.go
````

Set up credentials for kafka and send your statement(temporary and will eventually not be necessary):

````
>>> SET kafka.key=JIM;
>>> SET kafka.secret=SECRET;
````

Send your statement:

````
>>> INSERT INTO `topic_1` SELECT * from `topic_0`;
````

And you should see something like this:

````
Statement successfully submited.
Statement ID: 2091b94c-0508-43eb-8052-5cf34d864279
Status: PENDING.
````


**If something doesn't work** or you need some help with trying out the demo, you can contact Jim Hughes (he has kindly offered himself to be mentioned here). Otherwise, your job is now submitted and will be running in seconds! 

You might want to delete the it using the id returned by the client, so we don't overload the service with jobs. Is this case you would do:

```
kubectl-ccloud-config get devel
export KUBECONFIG=${HOME}/.kube/ccloud-config/devel/kubeconfig
kubectl config use-context k8s-4m2mf
kubectl port-forward -n fcp-system service/apiserver 8080:80 # set up port forward

curl --location --request DELETE 'http://localhost:8080/apis/sql/v1alpha1/orgs/org/environments/env/sqljobs/2091b94c-0508-43eb-8052-5cf34d864279' \
--header 'Accept: application/json'
```

Please note that there is no guarantee that this will be up and running at all times. If you want to change the parameters used to initialize the client (like gateway address, environment, Kafka cluster, extra properties, and so on), you can edit them in the [demo_devel.go](./_examples/devel/demo_devel.go) file.


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
