package main

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	// "time"
	// "sync/atomic"
)

func listen() {

	// listen on socket
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		slog.Error("failed to bind tcp listener", "addr", ":6379", "error", err)
		os.Exit(1)
	}
	slog.Info("tcp listener ready", "addr", ":6379")

	defer l.Close()
	for {

		conn, err := l.Accept()
		if err != nil {
			slog.Error("failed to accept incoming connection", "error", err)
			os.Exit(1)
		}

		slog.Info("client connected", "remote_addr", conn.RemoteAddr().String())

		go handle(conn)
	}

}

func handle(conn net.Conn) {

	remoteAddr := conn.RemoteAddr().String()
	slog.Debug("starting client handler", "remote_addr", remoteAddr)

	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {

		byteType, err := readValue(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Debug("client disconnected", "remote_addr", remoteAddr)
			} else {
				slog.Warn("failed to read RESP type marker", "remote_addr", remoteAddr, "error", err)
			}
			return
		}

		if byteType == '*' {
			_, err = parseArray(reader)
			if err != nil {
				slog.Warn("failed to parse RESP array", "remote_addr", remoteAddr, "error", err)
				return
			}
		} else if byteType == '$' {
			_, err := parseBulk(reader)
			if err != nil {
				slog.Warn("failed to parse RESP bulk string", "remote_addr", remoteAddr, "error", err)
				return
			}
		} else {
			slog.Warn("unsupported RESP type marker", "remote_addr", remoteAddr, "type", string(byteType))
		}

		_, err = conn.Write([]byte("+OK\r\n"))
		if err != nil {
			slog.Warn("failed to write response to client", "remote_addr", remoteAddr, "error", err)
			return
		}

	}

}

// Reads first byte from stream to determine data type
func readValue(reader *bufio.Reader) (byte, error) {

	byteType, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}

	return byteType, nil
}

// Parses an array message
func parseArray(reader *bufio.Reader) (string, error) {

	slog.Debug("decoding RESP array header")
	msg, err := reader.ReadString('\n')
	if err != nil {
		slog.Warn("failed to read RESP array length", "error", err)
		return "", err
	}

	length, err := strconv.Atoi(strings.TrimRight(msg, "\r\n"))
	if err != nil {
		slog.Warn("invalid RESP array length", "raw", strings.TrimRight(msg, "\r\n"), "error", err)
		return "", err
	}
	slog.Debug("decoded RESP array length", "length", length)

	for i := 0; i < length; i++ {

		byteType, err := readValue(reader)
		if err != nil {
			slog.Warn("failed to read RESP element type", "index", i, "error", err)
			return "", err
		}
		if byteType == '$' {
			msg, err := parseBulk(reader)
			if err != nil {
				slog.Warn("failed to parse RESP bulk element", "index", i, "error", err)
				continue
			}
			slog.Debug("decoded RESP bulk element", "index", i, "value", msg)
		} else {
			slog.Warn("unsupported RESP element type in array", "index", i, "type", string(byteType))
			continue
		}
	}

	return msg, nil
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
