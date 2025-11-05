package main

import (
	"fmt"
	"net"
	"log"
	"my.http/internal/request"
)


func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		log.Print("hello","hello")

		r, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		rl := r.RequestLine
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			   rl.Method, rl.RequestTarget, rl.HttpVersion)
	}

	
}

