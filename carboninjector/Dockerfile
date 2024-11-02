FROM alpine:3.7
RUN apk --no-cache add python3 tini
RUN pip3 install pika
ADD carboninjector.py /
ENTRYPOINT ["/sbin/tini", "--", "python3", "/carboninjector.py"]
