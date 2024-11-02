FROM golang:1.10.1-alpine as builder
WORKDIR /go/src/snmpmon
COPY ./snmpmon.go /go/src/snmpmon/snmpmon.go
RUN apk --no-cache add git
RUN go get github.com/streadway/amqp
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o build/snmpmon snmpmon.go

FROM alpine:3.7
COPY --from=builder /go/src/snmpmon/build/snmpmon /snmpmon
RUN apk --no-cache add tini
ENTRYPOINT ["/sbin/tini", "--", "/snmpmon"]

