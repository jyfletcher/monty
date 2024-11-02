# CarbonInjector

CarbonInjector is designed to pull from an AMQP 0-9-1 queue (eg RabbitMQ) and inject the pulled messages into Carbon, which is the input daemon for Graphite.

## Runtime
Configuration is through environment variables passed into the container at runtime:
```
export CARBON_HOST=host.name.tld
export CARBON_PORT=7003
export AMQP_CONNECTION_STRING=amqp://user:pass@host.domain.tld:5672/vhost
export QUEUE=carbon
```

## Build

The build script just creates a docker image tagged with the current date.


