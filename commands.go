package main

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

var acceptedCommands = []string{
	"ping", "set", "get", "hset", "hget", "hgetall"}

func dispatchCommand(fields []string, kvStore kvStore) ([]byte, error) {

	if len(fields) == 0 {
		slog.Warn("received empty command")
		return nil, fmt.Errorf("ERR empty command")
	}

	command := strings.ToLower(fields[0])
	if !slices.Contains(acceptedCommands, command) {
		slog.Warn("received unsupported command", "command", command, "argument_count", len(fields)-1)
		return nil, fmt.Errorf("ERR unsupported command: %s", command)
	}
	slog.Debug("dispatching command", "command", command, "argument_count", len(fields)-1)

	var resp []byte
	switch command {
	case "ping":
		if len(fields) > 2 {
			slog.Warn("received ping with too many arguments", "argument_count", len(fields)-1)
			return nil, fmt.Errorf("ERR wrong number of arguments for 'ping' command")
		} else if len(fields) == 1 {
			resp = []byte("+PONG\r\n")
		} else {
			resp = []byte(
				"$" + strconv.Itoa(len(fields[1])) + "\r\n" + fields[1] + "\r\n",
			)
		}
	case "set":
		if len(fields) != 3 {
			slog.Warn("received ping with wrong number of arguments", "argument_count", len(fields)-1)
			return nil, fmt.Errorf("ERR wrong number of arguments for 'set' command")
		}
		key := fields[1]
		value := fields[2]
		kvStore.Store[key] = value
		resp = []byte("+OK\r\n")
	case "get":
		if len(fields) != 2 {
			slog.Warn("received ping with wrong number of arguments", "argument_count", len(fields)-1)
			return nil, fmt.Errorf("ERR wrong number of arguments for 'get' command")
		}

		key := fields[1]
		value, ok := kvStore.Store[key]
		if !ok {
			slog.Warn("key not found", "key", key)
			return nil, fmt.Errorf("ERR key not found for 'get' command")
		}
		resp = []byte(
			"$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n",
		)

	}
	slog.Debug("command completed", "command", command, "response_bytes", len(resp))

	return resp, nil

}
