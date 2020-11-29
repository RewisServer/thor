# tHor

Thor is a Prometheus push and aggregation gateway, which allows volatile jobs to push their metrics to a gateway instead of being a scrape target itself.

# Goals

Currently the [Pushgateway](https://github.com/prometheus/pushgateway) is the number one choice when it comes to pushing metrics into Prometheus. The problem with that is, that it is only for single run jobs to expose their metric once.

In our case we wanted to have a gateway, which of course allows pushing as well, but which also allows the clients to push multiple metrics at multiple times during a job. This comes in handy when we use it for our game servers observing metrics like player counts, command uses, etc.

Also, this is more or less a fork of the Pushgateway, as we wanted to have it 100% compatible with Prometheus clients.

# Usage

You can either download and build the sources yourself by using the `build.sh` script.

Or you could just use Docker and run the image, with:

```sh
docker run --expose 9091 volixug/thor:0.2.1
```

After that you can use a Prometheus client of your choice (e.g. [client_java](https://github.com/prometheus/client_java)) and connect to it with:

```java
PushGateway gateway = new PushGateway("localhost:9091");
```
