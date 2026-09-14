# Redis Sampling State Extension Changelog

## Unreleased

- Initial scaffold: `redis_sampling_state` extension implementing the counter
  store contract (`AddCounts`/`ReadCounts`) with pipelined `HINCRBYFLOAT`,
  `HGETALL`, cluster hash-tagged keys, and per-bucket TTL.
