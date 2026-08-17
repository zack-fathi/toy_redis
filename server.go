package main

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
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
		resp := []byte("+OK\r\n")
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Debug("client disconnected", "remote_addr", remoteAddr)
			} else {
				slog.Warn("failed to read RESP type marker", "remote_addr", remoteAddr, "error", err)
			}
			return
		}

		if byteType == '*' {
			fields, err := parseArray(reader)
			if err != nil {
				slog.Warn("failed to parse RESP array", "remote_addr", remoteAddr, "error", err)
				return
			}

			resp, err = handleCommands(fields)
			if err != nil {
				slog.Warn("command failed", "remote_addr", remoteAddr, "error", err)
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

		_, err = conn.Write(resp)
		if err != nil {
			slog.Warn("failed to write response to client", "remote_addr", remoteAddr, "error", err)
			return
		}
		slog.Debug("response written", "remote_addr", remoteAddr, "bytes", len(resp))

	}

}
