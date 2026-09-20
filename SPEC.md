# Toy Redis Specification

## Goal

Build a small Redis-compatible in-memory data store in Go. The project is for learning about TCP servers, RESP, concurrent shared state, and append-only persistence. It is not intended to implement all of Redis.

## Commands

The server supports:

- `PING` with an optional message
- `SET key value`
- `GET key`
- `HSET hash field value`
- `HGET hash field`
- `HGETALL hash`

Command names are case-insensitive.

## RESP

The server must decode client command arrays containing bulk strings and encode these response types:

- Simple strings
- Bulk strings
- Null bulk strings
- Errors
- Integers
- Arrays

Every RESP line ends with `\r\n`. Bulk string payloads are read according to their declared byte length.

## Storage

The database contains:

```text
strings: key -> value
hashes:  hash -> field -> value
```

The database is shared by all client connections. String operations use one `sync.RWMutex`; hash operations use another. All map reads and writes must use the appropriate lock.

## Server

The server listens on port `6379` and keeps client connections open for multiple commands. Each connection is handled in its own goroutine. A client error response should not close the connection; malformed input, connection failures, and unrecoverable persistence errors may end a connection.

## AOF Persistence

Mutating commands are stored in `database.aof` as RESP command arrays:

- `SET`
- `HSET`

Read-only commands are not persisted:

- `PING`
- `GET`
- `HGET`
- `HGETALL`

On startup, the server opens or creates the AOF, replays each saved command into the new database, and only then starts accepting clients. Replay must not append the replayed commands again.

AOF writes are protected by a mutex. The file should be synchronized periodically and closed during shutdown.

## Testing

Tests should cover:

- RESP decoding and encoding
- Command behavior and RESP responses
- Missing keys and wrong argument counts
- Concurrent database access with the race detector
- Concurrent TCP clients using shared database and AOF state
- AOF replay after a restart

Run the test suite with:

```sh
go test -race ./...
```

## Completion Criteria

The project is complete when `redis-cli` can connect, the listed commands work, multiple clients can safely share state, and `SET`/`HSET` data survives a restart through AOF replay.
