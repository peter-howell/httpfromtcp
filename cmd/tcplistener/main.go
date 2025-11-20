package main

import (
	"fmt"
	"net"

	"github.com/peter-howell/httpfromtcp/internal/request"
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

		r, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		fmt.Printf("%s\n", r)
	}


}

