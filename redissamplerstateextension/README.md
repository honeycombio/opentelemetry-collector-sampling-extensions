# Redis Sampler State Extension

Type: `redis_sampler_state`

Backs the `adaptive_tail_sampling` processor's `shared_counters` option with
Redis. Every collector instance publishes its per-interval traffic counts with
pipelined `HINCRBYFLOAT` and reads back the merged totals, so
`goal_throughput` becomes a budget shared by the whole fleet.

> [!WARNING]
> Experimental. Configuration may change without notice.

## Configuration

```yaml
extensions:
  redis_sampler_state:
    endpoint: redis:6379        # required, host:port
    password: ${env:REDIS_PW}   # optional
    db: 0                       # optional, Redis logical database
    key_prefix: samplerstate    # optional, default samplerstate
    bucket_ttl: 10m             # optional, how long interval buckets persist

processors:
  adaptive_tail_sampling:
    rules:
      - name: default
        sampler:
          type: adaptive_throughput
          goal_throughput: 1000            # fleet-wide spans/sec
          fingerprint_attributes:
            - resource.attributes["service.name"]
          shared_counters:
            extension: redis_sampler_state

service:
  extensions: [redis_sampler_state]
```

All instances sharing a Redis deployment must use the same `key_prefix`, rule
names, goals, and sampler intervals; each instance recomputes rates locally
from the merged counts.

## Data model

One Redis hash per sampler per adjustment interval:

```
<key_prefix>:{<samplerID>}:<bucket> -> { <sampling key>: <count>, ... }
```

The samplerID is a Redis Cluster hash tag, keeping a sampler's buckets in one
slot. Buckets expire after `bucket_ttl`; the processor only reads the bucket
it just wrote, so the TTL merely needs to outlive the adjustment interval.
