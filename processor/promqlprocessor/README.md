# Promql Processor

| Status                   |               |
| ------------------------ |---------------|
| Stability                | [development] |
| Supported pipeline types | metrics       |
| Distributions            | [contrib]     |

Processes data using [promql](https://prometheus.io/docs/prometheus/latest/querying/basics/).

Currently supports metric types:
- Gauge
- Sum

Don't replace metric name (whose label is `__name__` ) with promql's `label_replace`. It won't be respected by this processor.

Timestamp will be replace by current timestamp.
It foucuses on processing data of current scrape (or current metric slice).

This proccessor won't `honor_timestamps` for the following reasons:

1. avoid storing huge historical data in memory or disk
2. metrics could be scrape from multiple nodes with time different

## Getting Started

Storage Configurations

```yaml
tsdb:
  storage_dir: /data # ramdisk or tmpfs on k8s
  max_samples: 999999 # max samples storage in tsdb
  query_timeout: 30ms # query timeout
  metrics: # metrics to be store in tsdb
    - a
    - kube_.*
```

Query Configurations

```yaml
queries:
- query_string: a * on(namespace, pod) group_left(namespace, pod) kube_pod_labels{} # query string
  output_metric_name: a_pod_labels # output metric name
```
