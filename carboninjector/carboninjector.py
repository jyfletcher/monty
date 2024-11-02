#!/usr/bin/env python3

import logging
import os
import pika
import socket
import time

logging.basicConfig(level=logging.INFO)

def main(carbon_host, carbon_port, amqp, queue):
    # Bootstrap Carbon
    carbon = CarbonTransport(carbon_host, carbon_port, amqp, queue)
    # Bootstrap Rabbit
    params = pika.URLParameters(amqp)
    connection = pika.BlockingConnection(params)
    channel = connection.channel()
    channel.basic_qos(prefetch_count=100)
    channel.basic_consume(carbon, queue)
    try:
        carbon.open()
        channel.start_consuming()
    except BaseException as be:
        print(be)
        channel.stop_consuming()
    carbon.close()
    connection.close()


class CarbonTransport(object):

    def __init__(self, carbon_host, carbon_port, amqp, queue):
        self.carbon_conn = (carbon_host, carbon_port)
        self.amqp = amqp
        self.queue = queue
        self.sock = None


    def __call__(self, channel, method_frame, header_frame, body):
        logging.debug(method_frame.delivery_tag)
        logging.debug(body)
        tag = method_frame.delivery_tag
        status = self._send(body)
        if status == True:
            logging.debug("Message %s sent successfully" % tag)
            channel.basic_ack(delivery_tag=tag)
        else:
            ## requeue might be best as a configurable variable
            logging.warn("Sending nack for %s" % tag)
            channel.basic_nack(delivery_tag=tag, requeue=True)

    def open(self):
        logging.info("Connecting to Carbon")
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.settimeout(20)
        self.sock.connect(self.carbon_conn)

    def close(self):
        logging.info("Closing connection to Carbon")
        self.sock.shutdown(socket.SHUT_RDWR)
        self.sock.close()

    def _send(self, body):
        status = True
        while len(body) > 0:
            sent = self.sock.send(body)
            body = body[sent:]
            if sent == 0:
                logging.warn("Could not sent message to carbon")
                status = False
        self.sock.sendall("\n".encode())
        return status
    

if __name__ == '__main__':
    try:
        carbon_host = os.environ['CARBON_HOST']
        carbon_port = int(os.environ['CARBON_PORT'])
        amqp = os.environ['AMQP_CONNECTION_STRING']
        queue = os.environ['QUEUE']
        main(carbon_host, carbon_port, amqp, queue)
    except KeyError as ke:
        print("The following environment variables must be set:")
        print("CARBON_HOST")
        print("CARBON_PORT")
        print("AMQP_CONNECTION_STRING")
        print("QUEUE")
