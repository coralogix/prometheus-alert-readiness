# Kafka

## The Challenge
Consider as an example: attempting to run Kafka on Kubernetes. When running
a Kafka cluster, producer and consumer groups for topics are split into
_partitions_ which are replicated across a certain number of Kafka nodes. In
Kubernetes, the concept of a "Kafka node" maps to a Pod. As such, we are faced
with a dilemma as to when it is safe to terminate a Pod. On one hand, we need
to ensure that producer and consumer groups are always available; if a
partition's available replicas reduce to zero, then Kafka will suffer data loss
in that partition, and data resilience will have been disrupted by the
Kubernetes cluster operation. This is unacceptable in production clusters.

Proposed solutions that attempt to solve this, like using a
`PodDisruptionBudget` or strimzi's [`KafkaRoller`][kafka-roller], implicitly
assume the continued availability of the underlying Kubernetes nodes. Readiness
checks are simply insufficient to express both that other pods in the
StatefulSet should not be terminated (i.e. so that the PodDisruptionBudget will
act) and that the Pod should continue to receive traffic (as it is still the
leader of record for other partitions that are not currently under-replicated).
As a result, a naive cluster operator draining production nodes can and will
cause data loss.

## The Solution
The [Prometheus exporter for Kafka][kafka-exporter] exposes partition health in
two metrics: `kafka_server_replicamanager_underreplicatedpartitions_value` and
`kafka_server_replicamanager_offlinereplicacount_value`. Writing a Prometheus
alert to notify when a cluster is unhealthy is then as simple as writing the
following two alerts:

```yaml
- alert: KafkaUnderreplicatedPartitions
  expr: sum(kafka_server_replicamanager_underreplicatedpartitions_value) by (job) > 0
  labels:
    severity: warning
  annotations:
    summary: Kafka cluster {{$labels.job}} has underreplicated partitions
    description: The Kafka cluster {{$labels.job}} has {{$value}} underreplicated partitions
- alert: KafkaOfflinePartitions
  expr: sum(kafka_server_replicamanager_underreplicatedpartitions_value) by (job) > 0
  labels:
    severity: critical
  annotations:
    summary: Kafka cluster {{$labels.job}} has offline partitions
    description: The Kafka cluster {{$labels.job}} has {{$value}} offline partitions
```

When either one of these alerts are firing, the `prometheus-alert-readiness`
pod will respond as `NotReady` and prevent the cluster tooling from draining
any Kafka nodes, and therefore preventing the cluster tooling from evicting
any additional Kafka pods. When the number of underreplicated partitions
returns to zero, i.e. all partitions in the Kafka cluster are fully replicated
again, then the `prometheus-alert-readiness` pod will respond as `Ready` and
allow the cluster tooling to proceed.

[kafka-roller]: https://github.com/strimzi/strimzi-kafka-operator/blob/9b2678d7f9f6b61e84ce30c9c922cd55072c984c/cluster-operator/src/main/java/io/strimzi/operator/cluster/operator/resource/KafkaRoller.java
[kafka-exporter]: https://github.com/danielqsj/kafka_exporter
