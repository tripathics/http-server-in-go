package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string][]string
	Body    []byte
}

type Client struct {
	Conn    net.Conn
	Pending []byte
}

func parseRequest(client *Client) (Request, error) {
	buf := make([]byte, 1024)

	req := Request{
		Headers: make(map[string][]string),
		Body:    make([]byte, 0),
	}

	last_line_wo_crlf := make([]byte, 0)

	requestLineReached := false
	headersEndReached := false
	body_start_idx := -1

	for {
		var n int
		var err error = nil
		if len(client.Pending) > 0 {
			n = copy(buf, client.Pending)
			client.Pending = append([]byte(nil), client.Pending[n:]...)
		} else {
			n, err = client.Conn.Read(buf)
			if err != nil {
				if n > 0 {
					client.Pending = append([]byte(nil), buf[:n]...)
				}
				return req, err
			}
		}
		if n == 0 && len(client.Pending) == 0 {
			return req, nil
		}

		b := append(last_line_wo_crlf, buf[:n]...)
		buffer_size := len(b)

		i, j := 0, 0
		for j+2 <= buffer_size && !headersEndReached {
			if !bytes.Equal(b[j:j+2], []byte("\r\n")) {
				last_line_wo_crlf = append([]byte(nil), b[i:j+2]...)
				j++
			} else {
				j += 2
				line := b[i:j]
				i = j

				last_line_wo_crlf = make([]byte, 0)

				// headers end
				if bytes.Equal(line, []byte("\r\n")) {
					headersEndReached = true
				}

				line_trimmed := bytes.Trim(line, "\r\n")
				if !requestLineReached {
					parts := bytes.Split(line_trimmed, []byte(" "))
					if len(parts) != 3 {
						panic("Malformed request line")
					}
					req.Method, req.Path, req.Version = string(parts[0]), string(parts[1]), string(parts[2])
					requestLineReached = true
				} else if !headersEndReached {
					parts := bytes.SplitN(line_trimmed, []byte(":"), 2)
					if len(parts) != 2 {
						panic("Malformed request header")
					}
					key, val := strings.ToLower(string(parts[0])), strings.Trim(string(parts[1]), " ")
					if req.Headers[key] != nil {
						req.Headers[key] = append(req.Headers[key], val)
					} else {
						req.Headers[key] = []string{val}
					}
				}
			}
		}

		if headersEndReached {
			body_start_idx = j

			content_lengths := req.Headers["content-length"]
			if len(content_lengths) > 0 {
				parsed64, err := strconv.ParseInt(content_lengths[0], 10, 32)
				if err != nil {
					panic(err)
				}

				content_length := int(parsed64)

				curr_body_len := len(req.Body)
				upper_bound := body_start_idx + content_length - curr_body_len

				if upper_bound <= buffer_size {
					req.Body = append(req.Body, b[body_start_idx:upper_bound]...)
					client.Pending = append([]byte(nil), b[upper_bound:]...)
					break
				} else {
					req.Body = append(req.Body, b[body_start_idx:buffer_size]...)
				}
			} else {
				if j < buffer_size {
					client.Pending = append([]byte(nil), b[j:]...)
				}
				break
			}
		}
	}
	return req, nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	client := Client{
		Conn:    conn,
		Pending: make([]byte, 0),
	}

	for {
		parsedReq, parse_err := parseRequest(&client)
		if parse_err == io.EOF {
			return // connection is finished
		}
		if parse_err != nil {
			panic(parse_err)
		}
		// pending = parsedReq.Pending
		data, err := json.MarshalIndent(parsedReq, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
		fmt.Println(string(parsedReq.Body))

		response := "HTTP/1.1 200 OK\r\n" +
			"content-type: text/plain\r\n" +
			"content-length: 11\r\n" +
			"\r\n" +
			"hello world"
		conn.Write([]byte(response))

		if parsedReq.Headers["connection"] != nil &&
			strings.ToLower(parsedReq.Headers["connection"][0]) == "close" {
			break
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server listening at :8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(conn)
	}
}
