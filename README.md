# prometheus-alert-readiness

Translates firing Prometheus alerts into a Kubernetes readiness path.

## Why?
By running this container in a singleton deployment with PriorityClass `system-cluster-critical`, we can prevent automated rolling-update tooling from proceeding (as they will not proceed when such high-priority pods are `NotReady`) when they shouldn't, due to firing Prometheus alerts when underlying databases are in e.g. underreplicated status.

## Configuration
Configured by environment variables:
| Variable | Default | Description |
| -------- | ------- | ----------- |
| `PROMETHEUS_ENDPOINT` | `http://localhost:9090` | The location of the Prometheus endpoint to send API requests to. |
| `PROMETHEUS_API_TIMEOUT` | `10` | How long the readiness check should wait for Prometheus to respond before timing out. |
| `PROMETHEUS_ALERT_SEVERITIES` | `critical,warning` | A comma-separated string of severities that will cause `prometheus-alert-readiness` to respond `NotReady`. |
| `KUBE_LIVENESS_PATH` | `/live` | The HTTP path on which the Kubernetes liveness probe will listen. |
| `KUBE_READINESS_PATH` | `/ready` | The HTTP path on which the Kubernetes readiness probe will listen. |
| `KUBE_PROBE_LISTEN_PORT` | `8080` | The HTTP port on which the `prometheus-alert-readiness` will listen. |

## Local dev
1. Run `docker build -t prometheus-alert-readiness .` to build the container
2. Expose the remote Prometheus host locally by running e.g. `kubectl -n monitoring port-forward svc/kube-prometheus 9090`
3. Run `docker run --rm --network host prometheus-alert-readiness:latest` to run the container locally
4. Run `curl -i localhost:8080/ready` to trigger a readiness check
