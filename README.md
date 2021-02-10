# prometheus-alert-readiness

Translates firing Prometheus alerts into a Kubernetes readiness path.

## Why?
By running this container in a singleton deployment with PriorityClass `system-cluster-critical`, we can prevent automated rolling-update tooling from proceeding (as they will not proceed when such high-priority pods are `NotReady`) when they shouldn't, due to firing Prometheus alerts when underlying databases are in e.g. underreplicated status.

## Local dev
1. Run `docker build -t prometheus-alert-readiness .` to build the container
2. Expose the remote Prometheus host locally by running e.g. `kubectl -n monitoring port-forward svc/kube-prometheus 9090`
3. Run `docker run --rm --network host prometheus-alert-readiness:latest` to run the container locally
4. Run `curl -i localhost:8080/ready` to trigger a readiness check
