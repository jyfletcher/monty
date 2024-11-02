# TCPmon

TCPmon is a TCP Port monitor that connects to a port and reports success or failure.

tis designed to run as a docker container but of course would run fine as a standalone service. Monitoring endpoint configuration is through a list of environment variables containing a host and port will poll these ports at a 1 minute interval (maybe this should be configurable).

The output is to an AMQP 0-9-1 endpoint (eg RabbitMQ)

The output format is in the Graphite format:
```
string.identifier.separated.by.dots (metric value) (timestamp in unix epoch format)
```

Note that the output value is 0 for port down and 10 for port up.  This is to give a little room for future options (slow connect, etc). 

## Runtime

Each poll will generate 1 message containing newline separated entries.  This message can the be dumped raw into Graphite's socket but instead it is placed in a queue where it is either transported to another cluser or is directly picked up by another process and injected.

Configuration is through environment variables passed into the container at runtime:
```
export AMQP_URL="amqp://user:pass@host.domain.tld:5672/vhost"
export AMQP_EXCHANGE="myexchange"
export AMQP_ROUTING_KEY="tcp.status"
export CHECK_HOST_1="host1.domain.tld:443"
export CHECK_HOST_2="host1.domain.tld:80"
export CHECK_HOST_3="host2.domain.tld:5672"
export TIMEOUT_SEC=5
# The timeout is how long we will wait, in seconds, for the socket to open before giving up and considering it closed
export OUTPUT_PREFIX="mynamespace"
export DEBUG="false"
```

## Build

The build script is just a simple script to build a docker image and have it tagged with the current date.
