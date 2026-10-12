# opentelemetry-collector-sampling-extensions

[![OSS Lifecycle](https://img.shields.io/osslifecycle?file_url=https%3A%2F%2Fraw.githubusercontent.com%2Fhoneycombio%2Fopentelemetry-collector-sampling-extensions%2Fmain%2FOSSMETADATA)](https://github.com/honeycombio/home/blob/main/honeycomb-oss-lifecycle-and-practices.md)

This repository hosts sampling-related extensions for the
[OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).

Adaptive samplers often need state or coordination that is normally private
to one collector instance, for example the live size of the fleet. The
extensions in this repository back that state with shared infrastructure so a
fleet of collector instances can sample as one coordinated system rather than
a set of independent ones.

> [!WARNING]
> Experimental. Interfaces and configuration may change without notice. Do not
> depend on this for critical workloads yet.

These extensions are expected to move into the OpenTelemetry Collector once
the capabilities they provide settle upstream. This repository is where they
are developed in the meantime, so the module path will change when that
happens. Pin a version if you adopt one, and expect to repoint the import
rather than to receive it at this path indefinitely.

## Extensions

| Extension | Type | Backend |
|-----------|------|---------|
| [redisfleettracker](./extension/redisfleettracker) | `redis_fleet_tracker` | Redis |

## The fleet tracker contract

`redisfleettracker` exposes live fleet membership to other components,
satisfied structurally so no shared package is needed:

```go
type fleetTracker interface {
	SubscribeMemberCount(callback func(count int)) (cancel func(), err error)
}
```

This is the contract consumed by the `adaptive_tail_sampling` processor, which
divides each `adaptive_throughput` rule's `goal_throughput` by the live member
count so the goal reads as a fleet-wide budget
([contrib#51544](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/51544)).

## License

[Apache 2.0](./LICENSE)
