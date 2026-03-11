# Prometheus Metrics & Grafana Dashboards

> prometheus/client_golang v1.23.2
> Grafana: latest (via Docker)
> Loki: latest (via Docker)

## Prometheus Go Client

### Metric Types

| Type | Use Case | Lurkarr Example |
|------|----------|-----------------|
| Counter | Monotonically increasing | `lurk_searches_total`, `errors_total` |
| Gauge | Can go up/down | `download_client_queue_size`, `paused` |
| Histogram | Distributions/latencies | `lurk_duration_seconds`, `http_request_duration_seconds` |
| Summary | Pre-calculated quantiles | (not used — histograms preferred) |

### Registration with promauto

```go
var LurkSearchesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
    Namespace: "lurkarr",
    Subsystem: "lurk",
    Name:      "searches_total",
    Help:      "Total number of lurk searches triggered.",
}, []string{"app_type", "instance"})
```

### Current Lurkarr Metrics

**Lurking:** searches_total, missing_found_total, upgrades_found_total, duration_seconds, errors_total
**Queue Cleaner:** items_removed_total, strikes_total, blocklist_additions_total, run_duration_seconds
**Download Client:** queue_size (gauge), speed_bytes_per_second (gauge), paused (gauge)
**Scheduler:** executions_total, duration_seconds, errors_total
**HTTP:** requests_total, request_duration_seconds, response_size_bytes, rate_limit_hits_total
**Autoimport:** runs_total, errors_total

### Naming Conventions (Prometheus best practices)

- `namespace_subsystem_name_unit` (e.g. `lurkarr_lurk_duration_seconds`)
- Use `_total` suffix for counters
- Use `_seconds` for durations
- Use `_bytes` for sizes
- Use `_info` for metadata gauges (value=1)

## Grafana Dashboard Design

### Professional Dashboard Patterns

1. **Row-based layout** — group panels by logical section
2. **Variables** at top: `$app_type`, `$instance`, `$interval`
3. **Stat panels** for KPIs (total searches, error rate, uptime)
4. **Time series** for trends (searches over time, queue sizes)
5. **Heatmaps** for latency distributions
6. **Table panels** for per-instance breakdowns
7. **Alert annotations** overlaid on graphs

### Dashboard JSON Structure

```json
{
  "annotations": { "list": [] },
  "editable": true,
  "panels": [...],
  "templating": {
    "list": [
      {
        "name": "app_type",
        "type": "query",
        "query": "label_values(lurkarr_lurk_searches_total, app_type)"
      }
    ]
  },
  "time": { "from": "now-24h", "to": "now" },
  "refresh": "30s"
}
```

### Planned Dashboards (10)

1. **Lurking** — search rates, missing/upgrade trends, errors, duration histograms per app/instance
2. **Queue Cleaner** — strikes, removals, blocklist, stalled/slow/failed breakdown
3. **Download Clients** — queue sizes, speeds, paused states per client
4. **Auto-Import** — runs, success/error, score distributions
5. **Scheduler** — task executions, durations, errors, timeline
6. **HTTP/API** — request rates, p50/p95/p99 latencies, error rates, rate limits
7. **Notifications** — send counts per provider, success/failure
8. **System** — Go runtime (goroutines, heap, GC), DB pool stats
9. **Loki Logs** — structured log exploration, error aggregation
10. **Arr Stack** — Sonarr/Radarr health, queue sizes from arr APIs

### PromQL Examples

```promql
# Lurk searches per hour
sum(rate(lurkarr_lurk_searches_total[1h])) by (app_type)

# P95 lurk duration
histogram_quantile(0.95, rate(lurkarr_lurk_duration_seconds_bucket[5m]))

# Error rate percentage
sum(rate(lurkarr_lurk_errors_total[5m])) / sum(rate(lurkarr_lurk_searches_total[5m])) * 100

# Current download speed
lurkarr_download_client_speed_bytes_per_second{direction="download"}

# HTTP request rate by status
sum(rate(lurkarr_http_requests_total[5m])) by (status)
```

## Loki Integration

- Scrape container stdout (slog JSON)
- LogQL queries: `{container="lurkarr"} | json | level="ERROR"`
- Dashboard variables: level filter, app_type, time range

## Monitoring Stack (deploy/docker-compose.monitoring.yml)

- Prometheus: port 9090, 30d retention
- Loki: port 3100
- Grafana: port 3000 (admin/lurkarr)
- Auto-provisioned datasources + dashboards via deploy/grafana/
