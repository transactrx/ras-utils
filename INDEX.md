# INDEX.md

Navigation map for LLM agents. Use this to find files by topic.

## Package Directory

| Package | Purpose | Entry Point |
|---------|---------|-------------|
| `rascache/` | TTL cache with singleflight | `cache.go` |
| `rasconfig/` | DB pools, env helpers, Snowflake | `db.go`, `environment.go` |
| `rasconversion/` | pgtype ↔ Go type conversions | `pgtypeHelpers.go` |
| `rasevents/` | NATS event publishing | `eventsHandler.go` |
| `rashttp/` | HTTP request/response helpers | `http.go` |
| `raslocation/` | Operating hours, timezone scheduling | `location.go` |
| `raslogging/` | HTTP logging middleware | `logging.go` |
| `rasstack/` | Middleware composition | `stack.go` |
| `rastime/` | TimeOfDay, DateRange, DayOfWeek | `time.go` |
| `rasvalidation/` | UUID, email, phone, NPI validators | `validation.go` |
| `rasworker/` | Worker pool with shutdown | `worker.go` |
| `rasauth/` | OAuth2 client credentials | `auth.go` |

## Dependency Graph

```
rastime ← raslocation
    ↑
(all others are independent)
```

## By Topic

| Need | Package | Key Functions |
|------|---------|---------------|
| Cache with TTL | `rascache` | `NewCache`, `GetOrStore` |
| DB connection pool | `rasconfig` | `InitDbPool`, `InitReadOnlyDbPool` |
| Snowflake connection | `rasconfig` | `NewSnowflakePool` |
| Env var with default | `rasconfig` | `GetEnvironmentVariableOrDefault` |
| Nullable → pgtype | `rasconversion` | `ConvertToPgtype*` |
| pgtype → Go | `rasconversion` | `ConvertFromPgtype*` |
| Publish NATS event | `rasevents` | `SendEvent`, `SendEventAsync` |
| JSON response | `rashttp` | `WriteJSON`, `OK`, `BadRequest` |
| Decode JSON body | `rashttp` | `DecodeJSON` |
| Check if location open | `raslocation` | `IsOpenAt`, `IsOpenAtZone` |
| Request logging | `raslogging` | `LoggingMiddleware` |
| Chain middleware | `rasstack` | `CreateStack` |
| Time without date | `rastime` | `TimeOfDay`, `NewTimeOfDay` |
| Date range | `rastime` | `DateRange`, `CalendarYear` |
| Validate UUID | `rasvalidation` | `IsValidUUID` |
| Validate NPI | `rasvalidation` | `IsValidNPI` |
| Worker pool | `rasworker` | `NewPool`, `Submit` |
| OAuth token | `rasauth` | `GetToken` |
