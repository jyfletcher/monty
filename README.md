# Monty

---

## Status

Note: This is a collection of code I have had lying around and used on and off for years. This repo will slowly grow as I clean up various pieces of it and publish.

## About

Monty is a collection programs that monitor various systems in a distributed, decoupled and scalable way. Metric storage is done with Graphite, transport of metrics is through RabbitMQ, triggering of alerts is done with Moira and dashboards are created in Grafana.

The concept is similar to Prometheus and similar systems but the usage of Graphite, where timestamps are included with the metric data, coupled with RabbitMQ, means that distributed systems can continue to collect and drop metrics in local queue servers independent of the central collection, reporting, and alerting system availability. When connection is restored the time-series data fills in at the appropriate times giving insight into the status of remote systems even during disconnects.
