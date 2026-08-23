package main

import (
	"bufio"
	"errors"

	// "fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

type aof struct {
	file  *os.File
	aofMu sync.Mutex
}

func (aof *aof) openAOF(path string) error {

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("failed to open or create AOF file", "path", path, "error", err)
		return err
	}

	aof.file = file
	return nil

}

func (aof *aof) append(fields []string) error {

	aof.aofMu.Lock()
	defer aof.aofMu.Unlock()

	encodedResp := encodeCommand(fields)
	bytesWritten, err := aof.file.Write(encodedResp)
	if err != nil {
		slog.Warn("failed to write AOF", "error", err)
		return err
	}
	if bytesWritten != len(encodedResp) {
		return io.ErrShortWrite
	}

	slog.Info("wrote AOF bytes", "bytes", bytesWritten)
	return nil
}

func (aof *aof) replay(db *database) error {
	if _, err := aof.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	reader := bufio.NewReader(aof.file)
	remoteAddr := "aof-replay"

	for {

		// Process your AOF text/RESP command here
		fields, err := decodeRequest(reader, remoteAddr)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			slog.Error("failed to replay AOF", "error", err)
			return err
		}

		_, err = dispatchCommand(fields, db, aof, false)
		if err != nil {
			slog.Error("failed to apply replayed command", "source", remoteAddr, "error", err)
			return err
		}
		// fmt.Print("Loaded log line: ", "line", line)
	}
}

func (aof *aof) sync() error {
	aof.aofMu.Lock()
	defer aof.aofMu.Unlock()

	if err := aof.file.Sync(); err != nil {
		slog.Warn("failed to sync AOF", "error", err)
		return err
	}
	return nil
}

func (aof *aof) close() error {
	aof.aofMu.Lock()
	defer aof.aofMu.Unlock()

	if aof.file == nil {
		return nil
	}
	err := aof.file.Close()
	if err != nil {
		slog.Warn("failed to close AOF file", "filename", aof.file.Name(), "error", err)
		return err
	}
	return nil
}
