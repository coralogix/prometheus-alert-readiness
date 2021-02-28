# Elasticsearch

## The Challenge
Consider as an example: attempting to run Elasticsearch on Kubernetes. When
running Elasticsearch, a given _shard_ will be replicated a certain number of
times between Elasticsearch nodes, to ensure that the shard will continue to
be available, even if a specific Elasticsearch node is unavailable. In
Kubernetes, the concept of an "Elasticsearch node" maps to a Pod. As such, we
are faced with a dilemma as to when it is safe to terminate a Pod. On one hand,
we need to ensure that each Elasticsearch shard always belongs to a Pod that is
ready; if a shard is replicated twice, but both Pods to which the shard belongs
are unavailable, then the shard itself will be unavailable, and the
availability of the Elasticsearch cluster will have been disrupted by the
Kubernetes cluster operation. This is unacceptable in production clusters.

Proposed solutions that attempt to solve this problem with a
`PodDisruptionBudget` (like Elastic's own [cloud-on-k8s] project)
are naive and insufficient. Elastic's official approach is to not report a
given Elasticsearch pod as ready until [the entire cluster is green][es-cluster-health].
However, if the cluster is momentarily yellow, then this results in the entire
cluster becoming unavailable, with cascading failures in dependent services
which are still functional, despite the cluster being in a yellow (i.e.
under-replicated, but not unavailable) state. The more mature approach is to
only check the health of the local Pod, i.e. to run `GET /_cluster/health?local=true`
as a readiness check, but this no longer couples Kubernetes's understanding of
readiness to Elasticsearch's notion of shard availability. Therefore, the fact
that specific Pods in Kubernetes are available or unavailable, and that
specific PodDisruptionBudgets are satisfied or unsatisfied, is no longer in
and of itself sufficient to safely signal to cluster tooling whether it is safe
to terminate the underlying Kubernetes nodes.

## The Solution
The [Prometheus exporter for Elasticsearch][es-exporter] exposes cluster health
information in a Prometheus metric called
`elasticsearch_cluster_health_status`. Writing a Prometheus alert to notify
when a cluster is unhealthy is then as simple as writing the following alert:

```yaml
alert: ElasticsearchClusterUnhealthy
expr: elasticsearch_cluster_health_status{color!="green"} != 0
labels:
  severity: warning
annotations:
  summary: ES cluster {{$labels.cluster}} is not healthy
  description: The ES cluster {{$labels.cluster}} is currently responding with color {{$labels.color}}.
```

When this alert is firing, the `prometheus-alert-readiness` pod will respond
as `NotReady` and prevent the cluster tooling from draining any Elasticsearch
nodes, and therefore preventing the cluster tooling from evicting any
additional Elasticsearch pods. When the cluster's health returns to green, then
the `prometheus-alert-readiness` pod will respond as `Ready` and allow the
cluster tooling to proceed.

[cloud-on-k8s]: https://github.com/elastic/cloud-on-k8s
[es-cluster-health]: https://github.com/elastic/helm-charts/blob/ffd109085023a37211c259302e2d076d84eeca94/elasticsearch/values.yaml#L228
[es-exporter]: https://github.com/justwatchcom/elasticsearch_exporter
