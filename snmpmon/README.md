# SNMPmon

SNMPmon is designed to run as a docker container and gather metrics from a local SNMP agent.  Since the SNMP agent is local, it can listen only on localhost, which circumvents some of the security concerns that some have with SNMP.  Of course the SNMP agent doesn't have to be local, but that is the intended use.

Output is to an SQS endpoint.

The output format is in the graphite format
```
string.identifier.separated.by.dots (metric value) (timestamp in unix epoch format)
```
## Runtime

Each poll will generate and queue 1 message containing newline separated entries.  This message can then be dumped raw into Graphite's TCP socket.

Configuration is through environment variables passed into the container at runtime:
```
export AWS_REGION="eu-west-1"
export SQS_QUEUE_URL="https://sqs.eu-west-1.amazonaws.com/000000000000/myqueue"
export SNMP_HOST=127.0.0.1
export SNMP_PORT=161
export SNMP_COMMUNITY=public
export OUTPUT_ID_PREFIX=int.aws.test.test.`hostname -s`
```

## Build

The build script creates a docker image tagged with the current date.
