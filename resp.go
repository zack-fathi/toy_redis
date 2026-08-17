package main

import (
	"bufio"
	"io"
	"log/slog"
	"strconv"
	"strings"
	// "slices"
)

// Reads first byte from stream to determine data type
func readValue(reader *bufio.Reader) (byte, error) {

	byteType, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}

	return byteType, nil
}

// Parses an array message
func parseArray(reader *bufio.Reader) ([]string, error) {

	slog.Debug("decoding RESP array header")
	msg, err := reader.ReadString('\n')
	if err != nil {
		slog.Warn("failed to read RESP array length", "error", err)
		return nil, err
	}

	length, err := strconv.Atoi(strings.TrimRight(msg, "\r\n"))
	if err != nil {
		slog.Warn("invalid RESP array length", "raw", strings.TrimRight(msg, "\r\n"), "error", err)
		return nil, err
	}
	slog.Debug("decoded RESP array length", "length", length)


	var fields []string
	for i := 0; i < length; i++ {

		byteType, err := readValue(reader)
		if err != nil {
			slog.Warn("failed to read RESP element type", "index", i, "error", err)
			return nil, err
		}
		if byteType == '$' {
			val, err := parseBulk(reader)
			if err != nil {
				slog.Warn("failed to parse RESP bulk element", "index", i, "error", err)
				return nil, err
			}
			slog.Debug("decoded RESP bulk element", "index", i, "value", val)
			fields = append(fields, val)
		} else {
			slog.Warn("unsupported RESP element type in array", "index", i, "type", string(byteType))
			continue
		}
	}

	return fields, nil
}

// Parses a bulk string message
func parseBulk(reader *bufio.Reader) (string, error) {

	slog.Debug("decoding RESP bulk string header")
	line, err := reader.ReadString('\n')
	if err != nil {
		slog.Warn("failed to read RESP bulk length", "error", err)
		return "", err
	}

	// extract length of string
	length, err := strconv.Atoi(strings.TrimRight(line, "\r\n"))
	if err != nil {
		slog.Warn("invalid RESP bulk length", "raw", strings.TrimRight(line, "\r\n"), "error", err)
		return "", err
	}

	// read full message into data
	data := make([]byte, length)
	_, err = io.ReadFull(reader, data)
	if err != nil {
		slog.Warn("failed to read RESP bulk payload", "length", length, "error", err)
		return "", err
	}

	// read CLRF
	_, err = reader.ReadString('\n')
	if err != nil {
		slog.Warn("failed to read RESP bulk trailing CRLF", "error", err)
		return "", err
	}

	return string(data), nil
}
