# Business Risk Reduction

## The Challenge
Performing a rolling restart of the Kubernetes cluster poses a small but
significant business risk. As a company that sells a log management service,
ultimately we cannot fully control the amount of log traffic that our customers
send us, and operators must be prepared to deal with sudden, unplanned surges
in traffic. At any given time, the cluster is prepared to deal with a certain
level of surge traffic. Due to database rebalancing, Kubernetes cluster
restarts can be thought of as using a certain amount of surge capacity. There
are two ways to deal with this need - to either increase resources during
cluster restarts, or to schedule cluster restarts when the likelihood of
requiring surge capacity is lower. As cluster restarts occur relatively
infrequently, the business make the decision to prefer the lower-cost approach.

The business has many times when it would independently like to make the
decision to temporarily "reduce risk", e.g. when large customers notify us of
planned traffic surges or when hosting important demos for
high-potential-value sales prospects, without coordinating with cluster
operators.

## The Solution
The business uses Jira Cloud to plan, prioritize, and schedule work. Atlassian
produces smartphone apps for Jira Cloud that make it easy and secure for
non-technical employees to view, create, and manipulate issues in Jira Cloud.

The [Prometheus exporter for Jira Cloud][jira-exporter] will expose Jira issues
as Prometheus metrics. By creating a "Business Events (BE)" Jira project, with
a simple workflow whereby issues are either "In Progress" or "Done", the
exporter can be configured with a JQL query to only expose new (In Progress)
issues within the Business Events project:

```
project = BE AND resolution is empty
```

Then the following Prometheus alert can be configured:

```yaml
alert: JiraBusinessEventInProgress
expr: sum(jira_cloud_issue) by (project) > 0
labels:
  severity: warning
annotations:
  summary: Business Event In Progress
  description: The business event {{$labels.key}} is in-progress
```

When a business executive wants to reduce risk, they can go into Jira Cloud and
create a Business Event issue. The existence of this issue causes the
Prometheus alert to fire, thus causing the `prometheus-alert-readiness` pod to
report `NotReady`, and thus preventing the cluster tooling from draining any
additional nodes. When the business event is over, the business executive can
go into Jira Cloud and move the issue to Done (i.e. resolved). This causes the
alert to stop firing, thus allowing the `prometheus-alert-readiness` pod to
resume reporting `Ready`, and thus permitting cluster tooling to proceed with
the cluster operation.

[jira-exporter]: (https://github.com/jwholdsworth/jira-cloud-exporter)
