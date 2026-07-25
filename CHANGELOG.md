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
- 2026-07-14T20:18:00+05:30 - treat Redis misses as clean cache misses
- 2026-07-15T09:13:00+05:30 - log cache read and write failures
- 2026-07-15T15:49:00+05:30 - add repository list users query
- 2026-07-16T10:32:00+05:30 - wire ListUsers through service
- 2026-07-16T18:05:00+05:30 - add GET users route
- 2026-07-17T08:44:00+05:30 - add user service tests
- 2026-07-17T14:31:00+05:30 - cover cache hit miss and singleflight behavior
- 2026-07-18T09:58:00+05:30 - add HTTP handler tests
- 2026-07-18T13:42:00+05:30 - cover duplicate email and unknown id responses
- 2026-07-19T11:19:00+05:30 - rename module to usersvc
- 2026-07-19T17:27:00+05:30 - refresh backend README after cleanup
- 2026-07-20T08:37:00+05:30 - add frontend package manifest
- 2026-07-20T11:25:00+05:30 - add frontend HTML and React entrypoints
- 2026-07-20T16:18:00+05:30 - build user management React UI
- 2026-07-20T21:54:00+05:30 - style responsive frontend dashboard
- 2026-07-21T09:42:00+05:30 - add frontend Docker and nginx config
- 2026-07-21T15:03:00+05:30 - add frontend service to compose
- 2026-07-22T10:46:00+05:30 - connect UI to live users list
- 2026-07-22T18:24:00+05:30 - add embedded worker runtime flag
- 2026-07-23T08:59:00+05:30 - add health endpoint
- 2026-07-23T12:41:00+05:30 - add Render deployment blueprint
- 2026-07-23T19:33:00+05:30 - add Vercel frontend rewrite
- 2026-07-24T09:16:00+05:30 - fix deployed API proxy target
- 2026-07-24T13:52:00+05:30 - pin Vercel framework build settings
- 2026-07-24T21:10:00+05:30 - document measured performance results
- 2026-07-25T08:20:00+05:30 - start verified project changelog

## Backend Bug Fixes
- Registered the pgx SQL driver for database/sql.
- Checked migration constructor errors and treated ErrNoChange as success.
- Fixed user lookup and delete SQL/route issues.
- Corrected RabbitMQ JSON content type and cleaned Go dependencies.
- 2026-07-25T09:05:00+05:30 - document backend bug fixes

## Backend Hardening
- Aligned Docker builder version and removed duplicate build stage.
- Added sentinel repository errors and driver-level duplicate email detection.
- Introduced service interfaces for repository, cache, and publisher dependencies.
- Added graceful API shutdown and improved worker failure handling.
- 2026-07-25T10:18:00+05:30 - document backend hardening work

## Cache And Listing
- Treated Redis nil responses as cache misses.
- Added warning logs for cache read/write failures.
- Added ListUsers in repository, service, and HTTP transport.
- 2026-07-25T11:42:00+05:30 - document cache and list endpoint work

## Tests
- Added service fakes for cache hit, miss, and singleflight behavior.
- Added handler stubs for duplicate email and not-found responses.
- Kept fakes aligned with cache-miss semantics.
- 2026-07-25T12:37:00+05:30 - document backend test coverage

## Frontend
- Created a Vite React app with user creation, lookup, list, and delete flows.
- Added responsive styling and Docker/nginx deployment support.
- Wired the UI to the real GET /users endpoint and refresh behavior.
- 2026-07-25T13:55:00+05:30 - document frontend buildout

## Repo Cleanup
- Renamed the module to usersvc and cleaned touched Go files.
- Added repository hygiene files and cleaned README claims.
- Preserved authorship cleanup and history maintenance notes outside code commits.
- 2026-07-25T15:04:00+05:30 - document repo cleanup and rename

## Deployment
- Added embedded worker support, Render blueprint, health checks, and frontend rewrites.
- Verified managed Postgres, Redis, RabbitMQ, Render, and Vercel configuration.
- Corrected Vercel root/framework settings and public API routing.
- 2026-07-25T16:26:00+05:30 - document deployment work

## Verification
- Ran build, vet, test, and race-test checks across backend changes.
- Rebuilt the Docker stack and verified create, duplicate, get, cache-hit, delete, and migration restart behavior.
- Confirmed live cloud end-to-end behavior after local Docker was off.
- 2026-07-25T17:48:00+05:30 - document verification runs

## Benchmarking
- Ran cached-read load testing with bombardier.
- Captured latency percentiles and rewrote performance notes with measured caveats.
- 2026-07-25T19:15:00+05:30 - document benchmarking results workflow
