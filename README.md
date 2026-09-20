# Toy Redis

A small Redis-inspired in-memory database written in Go. This project implements a focused subset of Redis while exploring TCP servers, RESP parsing, concurrency, and append-only persistence.

## Supported Commands

- `PING`
- `SET key value`
- `GET key`
- `HSET hash field value`
- `HGET hash field`
- `HGETALL hash`

Command names are case-insensitive.

## How It Works

A client sends a RESP array containing bulk strings. The server decodes the request into command fields, dispatches the command, updates the shared database when necessary, and writes a RESP response back to the client.

The main request path is:

```text
TCP connection
  -> RESP decoder
  -> command dispatcher
  -> database
  -> RESP response
```

### Storage

The database keeps two maps:

```text
strings: key -> value
hashes:  hash -> field -> value
```

String operations and hash operations use separate `sync.RWMutex` values. The same database pointer is shared by all client handler goroutines.

### AOF Persistence

`SET` and `HSET` mutations are appended to `database.aof` as RESP command arrays. On startup, the server opens or creates the file and replays the saved commands before accepting clients. Replay changes the database without appending the commands again.

The AOF has its own mutex so concurrent clients write complete RESP records without interleaving them.

## Project Files

- `main.go`: configures logging, creates the database, restores the AOF, and starts the server.
- `server.go`: listens for TCP connections and runs one handler goroutine per client.
- `resp.go`: decodes RESP command arrays and encodes command arrays for the AOF.
- `commands.go`: validates commands, dispatches handlers, and coordinates AOF persistence.
- `store.go`: owns the database maps, locks, and storage operations.
- `aof.go`: opens, replays, appends, syncs, and closes the AOF.
- `commands_test.go`: tests command behavior and RESP responses.
- `store_test.go`: exercises concurrent database access.
- `server_test.go`: exercises concurrent clients with `net.Pipe`.
- `SPEC.md`: concise project requirements.

## Running the Server

From the project directory:

```sh
go run .
```

The server listens on port `6379`. In another terminal, connect with:

```sh
redis-cli -p 6379
```

Example session:

```text
SET name Zackery
GET name
HSET users zackery engineer
HGET users zackery
HGETALL users
```

To build and run a binary:

```sh
go build -o toy_redis .
./toy_redis
```

## Testing

Run all tests with the race detector:

```sh
go test -race ./...
```

Run a specific test:

```sh
go test -race -run TestConcurrentTCPRequests -count=1
```

The tests cover command responses, missing values, wrong argument counts, concurrent database access, concurrent TCP clients, and shared AOF writes.

## Resources

This project was developed while studying:

- [Build Redis from Scratch](https://www.build-redis-from-scratch.dev/), especially the server, RESP reader/writer, command, and AOF sections.
- [Redis protocol specification](https://redis.io/docs/latest/develop/reference/protocol-spec/), for RESP framing and data types.
- [Redis commands](https://redis.io/commands/), for command behavior and response conventions.
- [Go `sync` package](https://pkg.go.dev/sync), for `RWMutex` and concurrency control.
- [Go race detector](https://go.dev/doc/articles/race_detector), for checking concurrent access.
- [Go `net` package](https://pkg.go.dev/net), including `net.Pipe` for in-memory connection tests.

The implementation is an independent learning project, organized around the concepts from these resources.
