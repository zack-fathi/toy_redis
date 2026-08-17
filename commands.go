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

func dispatchCommand(fields []string) ([]byte, error) {

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
	}
	slog.Debug("command completed", "command", command, "response_bytes", len(resp))

	return resp, nil

}
