# Integration test

Wires the `redis_fleet_tracker` extension to the `adaptive_tail_sampling`
processor in opentelemetry-collector-contrib that consumes it, and checks the
pairing end to end against a real Redis.

This exists because the two sides are deliberately decoupled. The processor
resolves the extension by method shape through `host.GetExtensions` rather
than importing it, so neither module's build fails if the contract drifts. A
mismatch is a silent no-op at runtime. This test is what catches it.

It is a separate module so the extension itself never depends on contrib.

```sh
docker run -d -p 6379:6379 redis:7-alpine
go test -tags integration ./...
```

`REDIS_ENDPOINT` overrides the default of `localhost:6379`.

The extension is required at its published version but `replace`d to the
working tree, so CI catches a change that breaks the contract before it is
released. Drop the replace to run against the published artifact instead.
Both modes pass.

The contrib processor is still pinned to a pseudo-version, because
`fleet_tracker` merged after contrib v0.162.0 and is not in a tagged release
yet. Move it to the release version once one ships.
