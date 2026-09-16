package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	requests := []byte(
		"GET /one HTTP/1.1\r\n" +
			"Host: localhost\r\n" +
			"\r\n" +
			"GET /two HTTP/1.1\r\n" +
			"Host: localhost\r\n" +
			"\r\n" +
			"GET /three HTTP/1.1\r\n" +
			"Host: localhost\r\n" +
			"\r\n" +
			"GET /four HTTP/1.1\r\n" +
			"Host: localhost\r\n" +
			"\r\n",
	)

	_, err = conn.Write(requests)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		fmt.Print(string(buf[:n]))
	}
}
