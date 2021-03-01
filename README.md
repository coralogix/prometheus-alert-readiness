# prometheus-alert-readiness
Exposes firing Prometheus alerts as a simple HTTP GET path, so that Kubernetes
can check the path as a [readiness probe].

[readiness probe]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes

## What challenge does this address?
There are two common classifications of Kubernetes clusters - those that are
self-restricted to only host stateless services, and those that attempt to run
stateful services as well. There are many benefits to deciding to run stateful
services in the same Kubernetes clusters that host stateless services, like
having a single control plane for the full system. However, there are also many
challenges.

One of the primary challenges of running stateful services (and databases) in
Kubernetes is that Kubernetes's methods of understanding workload availability
(i.e. Pod readiness probes) are insufficiently expressive for databases. As a
result, tooling that relies upon these generic methods to safely perform
cluster operations (e.g. rolling upgrades of Kubernetes nodes) are unable to
perform those operations safely, as they rely upon tooling which fails to
provide the complete picture as to whether it is safe or not to perform the
operation. Unfortunately, most cluster operations therefore involve human
operator intervention, which is time-consuming and error-prone, particularly
when human operators are responsible for multiple large clusters.

## How does this address the challenge?

### Collecting information
As Prometheus has gained traction as a monitoring solution, particularly among
Kubernetes users, it has become increasingly common for databases and other
stateful services to expose Prometheus-formatted metrics, especially
health-relevant metrics. As such, Prometheus is often configured to fire alerts
when these metrics produce values that are outside of normal, healthy ranges.

### Governing cluster operations
Most cluster tooling that attempts to govern cluster operations like rolling
node upgrades already pause progress when critical Kubernetes cluster
components, like the CNI, are unavailable.

The `prometheus-alert-readiness` Pod will report `NotReady` whenever the
configured Prometheus instance has alerts of the configured severity which are
currently firing. Therefore, by running this in a singleton Deployment with
`PriorityClass` `system-cluster-critical`, we can prevent automated
rolling-update tooling that already pauses when `system-cluster-critical` Pods
are `NotReady`, from proceeding with cluster operations when they shouldn't.

As a result, the `prometheus-alert-readiness` Deployment allows for completely
automated cluster upgrades, including stateful clusters, with broad
flexibility. Because it is not coupled to any individual database, it can be
used to safeguard any database (that exports Prometheus metrics) from cluster
operations. Business requirements that dictate the temporary suspension of
cluster operations can be accommodated by writing a Prometheus metrics exporter
and alert to systematically expose the business requirement, while ensuring
that business requirements are automatically respected and decoupled from the
requirement to otherwise coordinate with cluster operators. Additional
`prometheus-alert-readiness` Deployments can be deployed if there are multiple
Prometheus installations in the Kubernetes cluster; the `NotReady` status of
a single `prometheus-alert-readiness` Pod is sufficient to pause cluster
operations.

## Case Studies
There are three primary cases where we use `prometheus-alert-readiness` to
safeguard Kubernetes clusters:

* [Elasticsearch][study-elasticsearch], a database where the availability of
replicas across Kubernetes Pods must be maintained to safeguard against data
loss
* [Kafka][study-kafka], a streaming datastore where the availability of
replicas across Kubernetes Pods must be maintained to safeguard against data
loss
* [Business risk reduction][study-risk], to temporarily reduce business risks
that are sometimes associated with cluster operations.

[study-elasticsearch]: (/docs/case-studies/elasticsearch.md)
[study-kafka]: (/docs/case-studies/kafka.md)
[study-risk]: (/docs/case-studies/business-risk.md)

## Installation
`prometheus-alert-readiness` is distributed as a Helm chart. Run e.g.:
```bash
helm repo add prometheus-alert-readiness https://coralogix.github.io/prometheus-alert-readiness/chart-repo
helm install prometheus-alert-readiness prometheus-alert-readiness/prometheus-alert-readiness
```

## Configuration
The following process-relevant (with ensuing environment variables) options
are available:
| Helm Value | Environment Variable | Default | Description |
| -------- | -------- | ------- | ----------- |
| `configuration.prometheusEndpoint` | `PROMETHEUS_ENDPOINT` | `http://localhost:9090` | The location of the Prometheus endpoint to send API requests to. |
| `configuration.prometheusApiTimeout`| `PROMETHEUS_API_TIMEOUT` | `10` | How long the readiness check should wait for Prometheus to respond before timing out. |
| `configuration.prometheusAlertSeverities`| `PROMETHEUS_ALERT_SEVERITIES` | `critical,warning` | A comma-separated string of severities that will cause `prometheus-alert-readiness` to respond `NotReady`. |
| `configuration.kubeLivenessPath`| `KUBE_LIVENESS_PATH` | `/live` | The HTTP path on which the Kubernetes liveness probe will listen. |
| `configuration.kubeReadinessPath`| `KUBE_READINESS_PATH` | `/ready` | The HTTP path on which the Kubernetes readiness probe will listen. |
| `configuration.kubeProbeListenPort`| `KUBE_PROBE_LISTEN_PORT` | `8080` | The HTTP port on which the `prometheus-alert-readiness` will listen. |

The following Kubernetes Deployment-specific (and subsequently Helm-specific)
options are available:
| Helm Value | Default | Description |
| -------- | -------- | -------- |
| `configuration.kubeReadinessTiming.failureThreshold` | `1` | The `failureThreshold` to set for the Kubernetes readiness probe |
| `configuration.kubeReadinessTiming.initialDelaySeconds` | `0` | The `initialDelaySeconds` to set for the Kubernetes readiness probe |
| `configuration.kubeReadinessTiming.periodSeconds` | `5` | The `periodSeconds` to set for the Kubernetes readiness probe |
| `configuration.kubeReadinessTiming.successThreshold` | `24` | The `successThreshold` to set for the Kubernetes readiness probe |
| `configuration.kubeReadinessTiming.timeoutSeconds` | `5` | The `timeoutSeconds` to set for the Kubernetes readiness probe |
| `configuration.kubePodPriorityClass` | `system-cluster-critical` | The `PriorityClass` to set for the `Deployment` |

## Local dev
1. Run `docker build -t prometheus-alert-readiness .` to build the container
2. Expose the remote Prometheus host locally by running e.g. `kubectl -n monitoring port-forward svc/kube-prometheus 9090`
3. Run `docker run --rm --network host prometheus-alert-readiness:latest` to run the container locally
4. Run `curl -i localhost:8080/ready` to trigger a readiness check
