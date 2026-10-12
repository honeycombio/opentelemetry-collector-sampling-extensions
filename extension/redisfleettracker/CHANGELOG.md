# Redis Fleet Tracker Extension Changelog

## Unreleased

## v0.1.0

First release.

- `redis_fleet_tracker` extension. Tracks live collector fleet membership in a
  Redis sorted set, one key per fleet, refreshed on a jittered heartbeat and
  expired by TTL. Members deregister on clean shutdown, and a dead fleet's key
  self-cleans.
- Exposes the fleet's live member count through `SubscribeMemberCount`, the
  contract the `adaptive_tail_sampling` processor resolves structurally to
  divide a fleet-wide `goal_throughput` by the instance count
  ([contrib#51544](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/51544)).
- `fleet` scopes which instances count together, defaulting to the extension's
  own component ID, so one Redis can back several environments.
- Authentication by `username` and `password`, with optional TLS. A Redis
  outage at startup does not block collector startup, and consumers keep the
  last good count until a heartbeat succeeds.
