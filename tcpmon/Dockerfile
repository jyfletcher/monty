FROM golang:1.9.2 as builder

WORKDIR /go/src/tcpmon
COPY ./tcpmon.go /go/src/tcpmon/tcpmon.go
RUN go get github.com/streadway/amqp

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -o build/tcpmon tcpmon.go

FROM alpine:3.7

RUN apk add tini

COPY --from=builder /go/src/tcpmon/build/tcpmon /tcpmon

ENTRYPOINT ["/sbin/tini", "--", "/tcpmon"]
