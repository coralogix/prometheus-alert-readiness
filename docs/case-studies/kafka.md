# Kafka

## The Challenge
Consider as an example: attempting to run Kafka on Kubernetes. When running
a Kafka cluster, producer and consumer groups for topics are split into
_partitions_ which are replicated across a certain number of Kafka nodes. In
Kubernetes, the concept of a "Kafka node" maps to a Pod. As such, we are faced
with a dilemma as to when it is safe to terminate a Pod. On one hand, we need
to ensure that producer and consumer groups are always available; if a
partition 

## The Solution
