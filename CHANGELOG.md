# Changelog

Reconstructed July timeline from verified project notes and repository state.

## Timeline
- 2026-07-01T09:12:00+05:30 - init repository ignores
- 2026-07-01T10:38:00+05:30 - add initial Go module files
- 2026-07-01T12:05:00+05:30 - add user domain model
- 2026-07-01T16:42:00+05:30 - add environment config loader
- 2026-07-02T08:56:00+05:30 - add migration filesystem
- 2026-07-02T11:21:00+05:30 - add database connection layer
- 2026-07-02T15:47:00+05:30 - add initial user repository
- 2026-07-03T09:18:00+05:30 - add Redis cache client
- 2026-07-03T13:52:00+05:30 - add RabbitMQ publisher and consumer
- 2026-07-04T10:14:00+05:30 - add user service layer
- 2026-07-04T17:36:00+05:30 - add HTTP handlers
- 2026-07-05T09:49:00+05:30 - add logging and recovery middleware
- 2026-07-05T15:22:00+05:30 - add API entrypoint
- 2026-07-06T10:03:00+05:30 - add worker entrypoint
- 2026-07-06T18:11:00+05:30 - add Docker runtime files
- 2026-07-07T14:28:00+05:30 - add initial backend README
- 2026-07-08T09:35:00+05:30 - register pgx SQL driver and migration handling
- 2026-07-08T16:09:00+05:30 - fix repository id lookup and delete semantics
- 2026-07-09T11:44:00+05:30 - clean Go dependencies
- 2026-07-09T19:30:00+05:30 - align Docker build with module version
- 2026-07-10T13:17:00+05:30 - tighten compose worker restart policy
- 2026-07-11T08:40:00+05:30 - add repository sentinel errors
- 2026-07-11T12:58:00+05:30 - introduce service interfaces
- 2026-07-12T10:26:00+05:30 - use errors.Is for service error handling
- 2026-07-12T18:06:00+05:30 - return precise HTTP conflict and not found responses
- 2026-07-13T09:51:00+05:30 - add graceful API shutdown
- 2026-07-13T16:34:00+05:30 - harden worker message failure handling
- 2026-07-14T11:07:00+05:30 - add worker RabbitMQ connection retry
