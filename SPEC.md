# Toy Redis — Project Specification

## 1. Project Goal

Build a small Redis-compatible in-memory data store from scratch in Go.

The goal is to reproduce the core functionality covered by the referenced tutorial while using an independent implementation and architecture.

The server should:

- Accept TCP client connections.
- Communicate using the Redis Serialization Protocol (RESP).
- Support a small set of Redis commands.
- Allow multiple clients to interact with shared state concurrently.
- Persist mutating commands using an Append Only File (AOF).
- Restore state by replaying the AOF when the server starts.

This project is not intended to implement the complete Redis feature set.

---

## 2. Scope

### Required Commands

The server must support:

- `PING`
- `SET`
- `GET`
- `HSET`
- `HGET`
- `HGETALL`

### Required RESP Types

The server must support the RESP types required by the commands above:

- Arrays
- Bulk strings
- Simple strings
- Errors
- Null bulk strings

Integer responses may also be implemented where useful.

### Required System Features

- TCP server
- Persistent client connections
- RESP decoding
- RESP encoding
- Command dispatch
- In-memory string storage
- In-memory hash storage
- Concurrent client handling
- Thread-safe shared state
- Append Only File persistence
- AOF replay during startup
- Periodic AOF synchronization

---

## 3. High-Level Architecture

The main request path should follow this structure:

```text
TCP Client
    |
    v
TCP Connection
    |
    v
RESP Decoder
    |
    v
Internal Request Representation
    |
    v
Command Dispatcher
    |
    v
Command Handler
    |
    v
Database / Storage
    |
    v
Internal Response Representation
    |
    v
RESP Encoder
    |
    v
TCP Client
```

Networking, protocol parsing, command execution, and storage should remain logically separated.

---

## 4. TCP Server Requirements

The server must:

- Listen on a configurable TCP address and port.
- Support Redis's conventional port `6379` by default.
- Continuously accept incoming TCP connections.
- Keep connections open for multiple commands.
- Detect client disconnects and close connection resources cleanly.
- Handle malformed input without crashing the entire server.

Each accepted connection should be handled independently so one slow client does not block other clients.

### Acceptance Criteria

- `redis-cli` can connect to the server.
- A client can send multiple commands over one connection.
- Two or more clients can connect at the same time.
- Disconnecting one client does not affect other clients.

---

## 5. RESP Decoder Requirements

The RESP decoder converts raw bytes from a TCP stream into an internal RESP value.

### Bulk Strings

The decoder must:

1. Detect the `$` type marker.
2. Read the declared byte length.
3. Consume the length delimiter `\r\n`.
4. Read exactly the declared number of bytes.
5. Consume the trailing `\r\n`.
6. Return the decoded value.

Bulk string contents must be read according to their declared byte length rather than by searching for newline characters.

A bulk string may contain arbitrary bytes, including newline characters.

### Arrays

The decoder must:

1. Detect the `*` type marker.
2. Read the declared element count.
3. Consume the count delimiter.
4. Parse exactly that many RESP values.
5. Return those values as a single RESP array.

Array parsing should reuse the general RESP parsing logic recursively.

### Stream Behavior

The decoder must not assume:

- one TCP read equals one command;
- one TCP packet contains a complete RESP value;
- multiple commands cannot arrive together.

The decoder must consume exactly one RESP value at a time from the stream.

### Acceptance Criteria

The decoder can convert a request equivalent to:

```text
SET name Zackery
```

into an internal representation equivalent to:

```text
[
    "SET",
    "name",
    "Zackery"
]
```

---

## 6. RESP Internal Representation

The application should define its own internal model for RESP values.

The representation should be capable of expressing at least:

- Bulk string
- Simple string
- Array
- Error
- Null

The representation should be usable by both:

- the RESP decoder;
- the RESP encoder.

The exact Go types and structures are an implementation decision.

---

## 7. RESP Encoder Requirements

The RESP encoder converts internal response values into RESP-formatted bytes.

It must support encoding:

### Simple String

Conceptually:

```text
+OK\r\n
```

### Bulk String

Conceptually:

```text
$5\r\nhello\r\n
```

### Null Bulk String

Conceptually:

```text
$-1\r\n
```

### Error

Conceptually:

```text
-ERR message\r\n
```

### Array

Conceptually:

```text
*N\r\n...
```

Each array element must itself be encoded as a RESP value.

### Acceptance Criteria

Responses produced by the server can be correctly interpreted by `redis-cli`.

---

## 8. Command Dispatcher Requirements

The dispatcher receives a decoded RESP request and determines which command should execute.

The dispatcher must:

- Verify that the request is a RESP array.
- Reject empty arrays.
- Treat the first element as the command name.
- Treat remaining elements as command arguments.
- Handle command names case-insensitively.
- Route supported commands to their handlers.
- Return an error for unsupported commands.
- Return an error when arguments are invalid.

The dispatcher should not contain TCP-specific logic.

Command handlers should not be responsible for parsing RESP byte streams.

---

## 9. `PING` Command

### Requirements

The server must support:

```text
PING
```

and return:

```text
PONG
```

The server should also support an optional message:

```text
PING hello
```

and return the provided message.

### Acceptance Criteria

From `redis-cli`:

```text
PING
```

returns:

```text
PONG
```

This command serves as the first end-to-end integration test of:

```text
TCP
-> RESP decoding
-> command dispatch
-> command execution
-> RESP encoding
-> TCP response
```

---

## 10. String Storage

The server must provide an in-memory string key/value store.

Conceptually:

```text
key -> value
```

The storage must be shared by all connected clients.

The storage must be safe for concurrent access.

---

## 11. `SET` Command

### Requirements

`SET` must:

- Accept a key and value.
- Store the value under the given key.
- Replace the previous value if the key already exists.
- Return a success response.
- Reject invalid argument counts.

Example:

```text
SET name Zackery
```

Response:

```text
OK
```

---

## 12. `GET` Command

### Requirements

`GET` must:

- Accept one key.
- Return the stored value when the key exists.
- Return RESP null when the key does not exist.
- Reject invalid argument counts.

Example:

```text
GET name
```

Response:

```text
Zackery
```

Missing key:

```text
GET missing
```

returns a null RESP value.

---

## 13. Hash Storage

The server must provide a second in-memory data structure representing Redis hashes.

Conceptually:

```text
hash-key
    |
    +-- field -> value
    +-- field -> value
    +-- field -> value
```

Hash storage must:

- be shared by all clients;
- be safe for concurrent access.

---

## 14. `HSET` Command

### Requirements

`HSET` must:

- Accept a hash key, field, and value.
- Create the hash if it does not exist.
- Insert or replace the specified field.
- Reject invalid argument counts.

Example:

```text
HSET users zackery engineer
```

Conceptual state:

```text
users:
    zackery -> engineer
```

---

## 15. `HGET` Command

### Requirements

`HGET` must:

- Accept a hash key and field.
- Return the field value if it exists.
- Return RESP null if the hash or field does not exist.
- Reject invalid argument counts.

---

## 16. `HGETALL` Command

### Requirements

`HGETALL` must:

- Accept one hash key.
- Return all field/value pairs for the hash.
- Encode the response as a RESP array.
- Return an empty array when appropriate.
- Reject invalid argument counts.

Conceptually:

```text
HGETALL users
```

may return values equivalent to:

```text
[
    "zackery",
    "engineer",
    "alice",
    "designer"
]
```

The exact ordering of map-backed results does not need to be guaranteed unless explicitly implemented.

---

## 17. Concurrency Requirements

The server must support multiple simultaneous clients.

Each connection should independently perform:

```text
read RESP request
-> dispatch command
-> execute command
-> encode response
-> write response
-> repeat
```

All connections share the same database state.

Shared state must be synchronized to avoid:

- data races;
- concurrent map access failures;
- inconsistent writes.

A slow or idle client should not prevent other clients from issuing commands.

### Acceptance Criteria

Two separate `redis-cli` sessions can:

1. connect simultaneously;
2. modify and read shared keys;
3. observe each other's updates.

The server should also pass Go's race detector for supported workloads.

---

## 18. Append Only File Persistence

The server must persist mutating commands using an Append Only File.

The AOF represents the sequence of operations required to reconstruct database state.

### Commands That Must Be Persisted

At minimum:

- `SET`
- `HSET`

### Commands That Must Not Be Persisted

Read-only operations should not be written to the AOF:

- `GET`
- `HGET`
- `HGETALL`
- `PING`

### Format

Commands should be written in RESP format so the existing RESP decoder can be reused when reading the file.

Conceptually:

```text
SET x 10
SET y 20
HSET users zackery engineer
```

becomes a sequence of RESP command arrays in the AOF.

---

## 19. AOF Replay

When the server starts, it must reconstruct its in-memory database from the AOF before accepting normal client traffic.

Startup flow:

```text
Start Server
    |
    v
Open AOF
    |
    v
Read RESP Command
    |
    v
Dispatch Command
    |
    v
Execute Command
    |
    v
Repeat Until EOF
    |
    v
Begin Accepting Clients
```

The replay path should reuse existing:

- RESP parsing;
- command dispatch;
- command handlers.

Commands executed during AOF replay must not be appended back into the AOF.

### Acceptance Criteria

1. Start the server.
2. Execute several `SET` and `HSET` commands.
3. Stop the server.
4. Restart the server.
5. Previously written values are available through `GET`, `HGET`, and `HGETALL`.

---

## 20. AOF Synchronization

The server should periodically request that buffered AOF writes are flushed to durable storage.

A reasonable default is approximately once per second.

The AOF subsystem must be safe when:

- multiple commands are being appended;
- a flush occurs while commands are being processed.

The implementation should avoid corrupting or interleaving AOF records.

---

## 21. Error Handling

The server should return appropriate RESP errors for situations such as:

- Unsupported command.
- Wrong number of arguments.
- Invalid RESP input.
- Invalid request shape.
- Unexpected RESP value type.
- Invalid bulk string length.
- Invalid array length.

Connection-specific parsing failures should not crash the entire server.

Fatal startup failures, such as being unable to bind the listening socket, may terminate the process.

---

## 22. Recommended Component Boundaries

A possible conceptual organization is:

```text
server
    TCP listener
    connection lifecycle

resp
    decoder
    encoder
    RESP value representation

commands
    dispatcher
    PING
    SET
    GET
    HSET
    HGET
    HGETALL

store
    string storage
    hash storage
    synchronization

aof
    append
    replay
    sync
```

These names and package boundaries are recommendations only.

The implementation should use an organization that keeps responsibilities separated.

---

## 23. Implementation Milestones

### Milestone 1 — TCP Server

Deliverables:

- TCP listener.
- Client connection handling.
- Persistent connections.
- Hard-coded valid RESP response.

Completion condition:

`redis-cli` can connect and receive a response.

---

### Milestone 2 — RESP Bulk Strings

Deliverables:

- Parse bulk string headers.
- Read exact byte lengths.
- Correctly consume trailing CRLF.

Completion condition:

Bulk strings with spaces and newline characters are handled correctly.

---

### Milestone 3 — RESP Arrays

Deliverables:

- Parse array lengths.
- Recursively decode array elements.
- Decode normal Redis command requests.

Completion condition:

A `SET name Zackery` request can be represented internally as three command elements.

---

### Milestone 4 — RESP Encoder

Deliverables:

- Simple strings.
- Bulk strings.
- Errors.
- Null values.
- Arrays.

Completion condition:

`redis-cli` correctly displays server responses.

---

### Milestone 5 — Command Dispatcher

Deliverables:

- Command extraction.
- Argument extraction.
- Case-insensitive lookup.
- Unknown command errors.
- Argument validation.

---

### Milestone 6 — `PING`

Deliverables:

- `PING`
- Optional message behavior.

Completion condition:

```text
redis-cli PING
```

returns:

```text
PONG
```

---

### Milestone 7 — Strings

Deliverables:

- String store.
- `SET`.
- `GET`.
- Missing-key null behavior.

---

### Milestone 8 — Hashes

Deliverables:

- Hash store.
- `HSET`.
- `HGET`.
- `HGETALL`.

---

### Milestone 9 — Concurrent Clients

Deliverables:

- Independent connection handlers.
- Shared synchronized database state.
- Race-free concurrent reads and writes.

---

### Milestone 10 — AOF Writing

Deliverables:

- AOF file management.
- Persist mutating commands.
- RESP-formatted command records.

---

### Milestone 11 — AOF Replay

Deliverables:

- Read AOF during startup.
- Parse commands using the existing RESP decoder.
- Replay mutations.
- Prevent replayed commands from being appended again.

Completion condition:

Database contents survive a server restart.

---

### Milestone 12 — Hardening

Deliverables:

- Invalid RESP handling.
- Wrong-argument errors.
- Unknown command errors.
- Connection cleanup.
- AOF error handling.
- Concurrency testing.
- Parser edge-case tests.

---

## 24. Suggested Testing Strategy

### RESP Unit Tests

Test:

- Empty bulk string.
- Normal bulk string.
- Bulk string containing spaces.
- Bulk string containing `\n`.
- Bulk string containing `\r\n`.
- Array containing multiple bulk strings.
- Multiple RESP messages in the same stream.
- RESP messages split across small reads.
- Malformed lengths.
- Missing CRLF.

### Command Unit Tests

Test:

- Correct command.
- Wrong argument count.
- Unknown command.
- Case-insensitive command names.
- Missing keys.
- Existing keys.
- Overwriting existing values.

### Integration Tests

Test complete TCP interactions:

```text
PING
SET
GET
HSET
HGET
HGETALL
```

### Concurrency Tests

Run multiple clients simultaneously performing reads and writes.

Run the server/tests with:

```text
go test -race ./...
```

### Persistence Tests

Test:

1. Execute mutations.
2. Stop the server.
3. Restart.
4. Confirm state was restored.

---

## 25. Non-Goals

Unless added later, this project does not need to implement:

- Redis Cluster.
- Replication.
- Pub/Sub.
- Transactions.
- Lua scripting.
- Streams.
- Lists.
- Sets.
- Sorted sets.
- Key expiration / TTL.
- Authentication.
- ACLs.
- RDB snapshots.
- AOF rewriting/compaction.
- Redis Sentinel.
- Redis modules.
- Full Redis protocol compatibility.

These features can be treated as future extensions.

---

## 26. Definition of Done

The project is complete when:

- `redis-cli` can connect to the server.
- Multiple clients can connect simultaneously.
- RESP requests are parsed correctly from a byte stream.
- RESP responses are encoded correctly.
- `PING`, `SET`, `GET`, `HSET`, `HGET`, and `HGETALL` work.
- Missing values return appropriate null responses.
- Shared data access is concurrency-safe.
- Mutating commands are recorded in an AOF.
- State survives server restart through AOF replay.
- Malformed requests return errors without crashing the server.
- Core parser, command, concurrency, and persistence behavior is covered by tests.
