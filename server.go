package main

import (
	"bufio"
	"log/slog"
	"net"
	"os"
)

func listen(kvStore kvStore) {

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

		go serverHandler(conn, kvStore)
	}

}

func serverHandler(conn net.Conn, kvStore kvStore) {

	remoteAddr := conn.RemoteAddr().String()
	slog.Debug("starting client handler", "remote_addr", remoteAddr)

	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {

		fields, err := decodeRequest(reader, remoteAddr)
		if err != nil {
			slog.Debug("request decoding ended", "remote_addr", remoteAddr, "error", err)
			return
		}

		resp, err := dispatchCommand(fields, kvStore)
		if err != nil {
			slog.Warn("command handling failed", "remote_addr", remoteAddr, "error", err)
			return
		}
		_, err = conn.Write(resp)
		if err != nil {
			slog.Warn("failed to write response to client", "remote_addr", remoteAddr, "error", err)
			return
		}
		slog.Debug("response written", "remote_addr", remoteAddr, "bytes", len(resp))

	}

}
