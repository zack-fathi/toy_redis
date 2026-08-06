package main

import (
	"bufio"
	// "fmt"
	"fmt"
	"net"

	// "time"
	// "sync/atomic"
	"log"
)

func listen() {

	// listen on socket
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()
	for {

		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handle(conn)
	}

}

func handle(conn net.Conn) {

	fmt.Println("Handling connection")

	// make a buffer
	// while there are still bytes to read
	// call read and put the bytes in the buffer
	// print out final message

	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {

		byteType := readValue(reader)

		var msg string
		if byteType == '*' {
			msg = parseArray(reader)
		} else if byteType == '$' {
			msg = parseBulk(reader)
		}

		fmt.Println(msg)

		_, err := conn.Write([]byte("+OK\r\n"))
		if err != nil {
			fmt.Println("Connection closed")
			return
		}

	}

}

// Reads first byte from stream to determine data type
func readValue(reader *bufio.Reader) byte {

	byte_type, err := reader.ReadByte()
	if err != nil {
		fmt.Println("Connection closed")
		return 'X'
	}

	return byte_type
}

// Parses an array message
func parseArray(reader *bufio.Reader) string {

	fmt.Println("You have read a array")
	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Connection closed")
		return ""
	}
	return msg
}

// Parses a bulk string message
func parseBulk(reader *bufio.Reader) string {

	fmt.Println("You have read a bulk string")
	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Connection closed")
		return ""
	}
	return msg
}
