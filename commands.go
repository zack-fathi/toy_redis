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

func dispatchCommand(fields []string, db *database) ([]byte, error) {

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
		dispatchPing(fields, &resp)
	case "set":
		dispatchSet(fields, &resp, db)
	case "get":
		dispatchGet(fields, &resp, db)
	case "hset":
		dispatchHset(fields, &resp, db)
	case "hget":
		dispatchHget(fields, &resp, db)
	case "hgetall":
		dispatchHgetall(fields, &resp, db)
	}
	slog.Debug("command completed", "command", command, "response_bytes", len(resp))

	return resp, nil

}

func dispatchPing(fields []string, resp *[]byte) {
	if len(fields) > 2 {
		slog.Warn("received 'ping' with too many arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for 'ping' command\r\n")
	} else if len(fields) == 1 {
		*resp = []byte("+PONG\r\n")
	} else {
		*resp = []byte(
			"$" + strconv.Itoa(len(fields[1])) + "\r\n" + fields[1] + "\r\n",
		)
	}
}

func dispatchSet(fields []string, resp *[]byte, db *database) {
	
	if len(fields) != 3 {
		slog.Warn("received 'set' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}
	db.setString(fields, resp)
}

func dispatchGet(fields []string, resp *[]byte, db *database) {
	

	if len(fields) != 2 {
		slog.Warn("received 'get' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}
	db.getString(fields, resp)
}

func dispatchHset(fields []string, resp *[]byte, db *database) {

	argCount := len(fields)
	if argCount < 4 || argCount%2 != 0 {
		slog.Warn("received 'hset' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}
	db.setHashFields(fields, resp)	
}

func dispatchHget(fields []string, resp *[]byte, db *database) {

	if len(fields) != 3 {
		slog.Warn("received 'hget' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}
	db.getHashField(fields, resp)
}

func dispatchHgetall(fields []string, resp *[]byte, db *database) {

	if len(fields) != 2 {
		slog.Warn("received 'hgetall' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}
	db.getAllHashFields(fields, resp)
}
