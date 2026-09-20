package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentTCPRequests(t *testing.T) {
	db := newTestDatabase()
	aofFile, err := os.CreateTemp(t.TempDir(), "toy-redis-*.aof")
	if err != nil {
		t.Fatal(err)
	}
	aofLog := &aof{file: aofFile}
	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for clientID := 0; clientID < 10; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			serverConn, clientConn := net.Pipe()
			defer clientConn.Close()

			handlerDone := make(chan struct{})
			go func() {
				serverHandler(serverConn, db, aofLog)
				close(handlerDone)
			}()

			reader := bufio.NewReader(clientConn)
			key := fmt.Sprintf("client-key-%d", id)
			value := fmt.Sprintf("client-value-%d", id)
			hash := fmt.Sprintf("client-hash-%d", id)
			field := fmt.Sprintf("client-field-%d", id)

			if err := writeCommandAndRead(clientConn, reader, []string{"SET", key, value}, "+OK\r\n"); err != nil {
				errs <- fmt.Errorf("client %d SET: %w", id, err)
				serverConn.Close()
				<-handlerDone
				return
			}
			if err := writeCommandAndRead(clientConn, reader, []string{"GET", key}, "$"+strconv.Itoa(len(value))+"\r\n"+value+"\r\n"); err != nil {
				errs <- fmt.Errorf("client %d GET: %w", id, err)
				serverConn.Close()
				<-handlerDone
				return
			}
			if err := writeCommandAndRead(clientConn, reader, []string{"HSET", hash, field, value}, ":1\r\n"); err != nil {
				errs <- fmt.Errorf("client %d HSET: %w", id, err)
				serverConn.Close()
				<-handlerDone
				return
			}
			if err := writeCommandAndRead(clientConn, reader, []string{"HGET", hash, field}, "$"+strconv.Itoa(len(value))+"\r\n"+value+"\r\n"); err != nil {
				errs <- fmt.Errorf("client %d HGET: %w", id, err)
				serverConn.Close()
				<-handlerDone
				return
			}

			clientConn.Close()
			<-handlerDone
		}(clientID)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if err := aofLog.close(); err != nil {
		t.Fatal(err)
	}
}

func writeCommandAndRead(conn net.Conn, reader *bufio.Reader, fields []string, want string) error {
	if _, err := conn.Write(encodeCommand(fields)); err != nil {
		return err
	}

	got, err := readRESPResponse(reader)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("response = %q, want %q", got, want)
	}
	return nil
}

func readRESPResponse(reader *bufio.Reader) (string, error) {
	typeByte, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	response := string(typeByte) + line

	if typeByte != '$' {
		return response, nil
	}

	length, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"))
	if err != nil {
		return "", err
	}
	if length < 0 {
		return response, nil
	}

	payload := make([]byte, length+2)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return "", err
	}
	return response + string(payload), nil
}
