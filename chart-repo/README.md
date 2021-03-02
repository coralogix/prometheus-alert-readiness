# Packaging directions
Run:
```bash
helm package ../chart
helm repo index . --url https://coralogix.github.io/prometheus-alert-readiness/chart-repo
```
