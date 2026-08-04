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

	fmt.Println("Handling connection\n")

	// make a buffer
	// while there are still bytes to read
	// call read and put the bytes in the buffer
	// print out final message

	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {

		_, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed")
			return
		}

		_, err = conn.Write([]byte("+OK\r\n"))
		if err != nil {
			fmt.Println("Connection closed")
			return
		}

	}

}