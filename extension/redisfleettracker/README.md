# Redis Fleet Tracker Extension

Type: `redis_fleet_tracker`

Tracks how many collector instances are currently alive in a "fleet" using a
Redis sorted set, and lets other components subscribe to the live member
count. Built for consumers that need fleet size to scale a per-instance
budget, most directly the `adaptive_tail_sampling` processor's per-key
rendezvous routing
([opentelemetry-collector-contrib#50577](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/50577)).

> [!WARNING]
> Experimental. Configuration and the `SubscribeMemberCount` contract may
> change without notice.

## Configuration

```yaml
extensions:
  redis_fleet_tracker:
    endpoint: redis:6379         # required, host:port
    password: ${env:REDIS_PW}    # optional
    db: 0                        # optional, Redis logical database
    tls:
      insecure: true             # optional, default true (plain TCP)
    fleet: ${env:DEPLOY_ENV}  # optional, see "Fleet scoping" below
    key_prefix: fleet_tracker # optional, default fleet_tracker
    heartbeat_interval: 3s    # optional, default 3s
    ttl: 10s                  # optional, default 10s, must exceed heartbeat_interval

service:
  extensions: [redis_fleet_tracker]
```

| Field | Default | Required | Description |
|-------|---------|----------|--------------|
| `endpoint` | - | yes | Redis server address, `host:port`. |
| `password` | `""` | no | Redis auth password. |
| `db` | `0` | no | Redis logical database number. |
| `tls` | `insecure: true` | no | TLS settings for the Redis connection. Disabled (plain TCP) by default; set `tls::insecure: false` to opt in. |
| `fleet` | the extension's own component ID | no | The group of instances counted together. See "Fleet scoping" below. |
| `key_prefix` | `fleet_tracker` | no | Namespaces every key this extension writes. |
| `heartbeat_interval` | `3s` | no | How often this instance refreshes its membership. |
| `ttl` | `10s` | no | How long a member is considered live after its last heartbeat. Must be greater than `heartbeat_interval`. |

## Fleet scoping

`fleet` defaults to the extension's own component ID, so instances with an
identical `redis_fleet_tracker` config count together automatically. Two
common patterns:

- Named instances, e.g. `redis_fleet_tracker/prod`, scope a fleet on their
  own via the component ID; no `fleet` setting needed.
- When identical configs across environments share one Redis backend, set
  `fleet` explicitly, e.g. `fleet: ${env:DEPLOY_ENV}`, so instances in
  different environments don't count each other.

Getting this wrong fails in one direction: fleets that should be separate
but collide inflate the observed count N, which under-samples everywhere
that count feeds a per-instance budget.

## Data model

One Redis sorted set per fleet, key `<key_prefix>:<fleet>`. Each heartbeat
tick pipelines:

```
ZADD    <key>  <now_ms + ttl_ms>  <member_id>   # register/refresh self
ZREMRANGEBYSCORE <key>  -inf  <now_ms>          # prune expired members
ZCARD   <key>                                    # live member count
PEXPIRE <key>  <2 * ttl_ms>                      # dead-fleet key self-cleans
```

Registering self happens before pruning, so self is always counted even if
its own previous heartbeat expired. `member_id` is a 16-character hex string:
the first 16 hex characters of `sha256(service.instance.id)` when that
resource attribute is set, otherwise a random UUID generated once at
startup. It is never derived from hostnames or pod names, which can be
reused across restarts and would then be indistinguishable from the
previous instance.

On shutdown, the extension makes a best-effort `ZREM <key> <member_id>` with
a short timeout before closing the Redis client.

## Failure semantics

- **Start-anyway**: `Start` attempts one synchronous heartbeat (~3s
  timeout) to prime the count, but returns `nil` regardless of whether it
  succeeds. A Redis outage at startup does not block collector startup;
  background heartbeats keep retrying.
- Consumers hold the last-good count until the next successful heartbeat;
  nothing is delivered to `SubscribeMemberCount` until a count is known.
- **Clock skew**: member expiry is computed from each instance's own clock.
  Skew beyond `ttl` can transiently prune a still-live member; it
  re-registers on its next heartbeat, so the effect is a brief undercount,
  not a stuck one.

## The SubscribeMemberCount contract

```go
type fleetTracker interface {
	SubscribeMemberCount(callback func(count int)) (cancel func(), err error)
}
```

Consumers assert this interface structurally against the extension
retrieved via `host.GetExtensions`; no shared package is required. The
callback fires once a count is known and again whenever it changes;
`cancel` unsubscribes and is idempotent, though one in-flight delivery may
still complete after it returns.

## Telemetry

See [documentation.md](./documentation.md) for the metrics this extension
emits.

## Installing with the OpenTelemetry Collector Builder

```yaml
extensions:
  - gomod: github.com/honeycombio/opentelemetry-collector-sampling-extensions/extension/redisfleettracker v0.1.0
```
