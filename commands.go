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
		dispatchPing(fields, &resp)
	case "set":
		dispatchSet(fields, &resp, kvStore)
	case "get":
		dispatchGet(fields, &resp, kvStore)
	case "hset":
		dispatchHset(fields, &resp, kvStore)
	case "hget":
		dispatchHget(fields, &resp, kvStore)
	case "hgetall":
		dispatchHgetall(fields, &resp, kvStore)
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

func dispatchSet(fields []string, resp *[]byte, kvStore kvStore) {
	if len(fields) != 3 {
			slog.Warn("received 'set' with wrong number of arguments", "argument_count", len(fields)-1)
			*resp = []byte("-ERR wrong number of arguments for command\r\n")
			return
		}
		key := fields[1]
		value := fields[2]
		kvStore.Store[key] = value
		*resp = []byte("+OK\r\n")
}

func dispatchGet(fields []string, resp *[]byte, kvStore kvStore) {
	if len(fields) != 2 {
		slog.Warn("received 'get' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}

	key := fields[1]
	value, ok := kvStore.Store[key]
	if !ok {
		slog.Warn("key not found", "key", key)
		*resp = []byte("$-1\r\n")
		return
	}
	*resp = []byte(
		"$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n",
	)
}

func dispatchHset(fields []string, resp *[]byte, kvStore kvStore) {

	argCount := len(fields)
	if argCount < 4 || argCount % 2 != 0{
		slog.Warn("received 'hset' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}

	hash := fields[1]
	if _, exists := kvStore.Hashes[hash]; !exists {
		kvStore.Hashes[hash] = make(map[string]string)
	}
	
	newFields := 0
	for i := 2; i < len(fields); i += 2 {
		field := fields[i]
		newValue := fields[i+1]

		if _, exists := kvStore.Hashes[hash][field]; !exists {
			newFields++
		}

		kvStore.Hashes[hash][field] = newValue
	}

	*resp = []byte(":" + strconv.Itoa(newFields) + "\r\n")
}

func dispatchHget(fields []string, resp *[]byte, kvStore kvStore) {

	if len(fields) != 3 {
		slog.Warn("received 'hget' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}

	hash := fields[1]
	field := fields[2]
	if _, exists := kvStore.Hashes[hash][field]; !exists {
		slog.Warn("key not found", "hash", hash, "field", field)
		*resp = []byte("$-1\r\n")
		return
    }
	
	value := kvStore.Hashes[hash][field]
	*resp = []byte(
		"$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n",
	)

}

func dispatchHgetall(fields []string, resp *[]byte, kvStore kvStore) {

	if len(fields) != 2 {
		slog.Warn("received 'hgetall' with wrong number of arguments", "argument_count", len(fields)-1)
		*resp = []byte("-ERR wrong number of arguments for command\r\n")
		return
	}

	hash := fields[1]
	if _, exists := kvStore.Hashes[hash]; !exists {
		slog.Warn("key not found", "key", hash)
		*resp = []byte("*0\r\n")
		return
    }
	
	var respArray []string
	for field, value := range kvStore.Hashes[hash] {
		respArray = append(respArray, field, value)
	}

	*resp = []byte("*" + strconv.Itoa(len(respArray)) + "\r\n")
	for _, value := range respArray {
		encodedValue := []byte(
			"$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n",
		)
		*resp = append(*resp, encodedValue...)
	}

}

