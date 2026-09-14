# opentelemetry-collector-samplingstate

[![OSS Lifecycle](https://img.shields.io/osslifecycle?file_url=https%3A%2F%2Fraw.githubusercontent.com%2Fhoneycombio%2Fopentelemetry-collector-samplingstate%2Fmain%2FOSSMETADATA)](https://github.com/honeycombio/home/blob/main/honeycomb-oss-lifecycle-and-practices.md)

Sampling state extensions for the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).

Adaptive samplers hold per-key state (traffic counts) that is normally private
to one collector instance. The extensions in this repository back that state
with shared infrastructure so a fleet of collector instances can sample
against combined budgets rather than per-instance ones.

> [!WARNING]
> Experimental. Interfaces and configuration may change without notice. Do not
> depend on this for critical workloads yet.

## Extensions

| Extension | Type | Backend |
|-----------|------|---------|
| [redissamplingstateextension](./redissamplingstateextension) | `redis_sampling_state` | Redis |

## The counter store contract

Extensions here implement the counter store interface consumed by the
`adaptive_tail_sampling` processor's `shared_counters` option
([opentelemetry-collector-contrib#50577](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/50577)),
satisfied structurally so no shared package is needed:

```go
AddCounts(ctx context.Context, samplerID string, bucket int64, counts map[string]float64) error
ReadCounts(ctx context.Context, samplerID string, bucket int64) (map[string]float64, error)
```

Counts are additive: every instance publishes what it observed for an interval
bucket, reads back the merged totals, and recomputes its own rates locally.

## License

[Apache 2.0](./LICENSE)
