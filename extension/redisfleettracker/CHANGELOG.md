# Redis Fleet Tracker Extension Changelog

## Unreleased

- Initial scaffold: `redis_fleet_tracker` extension. Tracks live fleet
  membership in a Redis sorted set (`ZADD`/`ZREMRANGEBYSCORE`/`ZCARD`/`PEXPIRE`
  per heartbeat) and exposes a `SubscribeMemberCount` callback for consumers
  that need to know the current fleet size.
