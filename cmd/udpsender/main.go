package main

import (
	"fmt"
	"net"
	"bufio"
	"log"
	"os"
)



func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	conn, err := net.DialUDP("udp",nil,  addr)
	if err != nil {
		log.Fatal("error", "error", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print(">")
		strIn, err := reader.ReadString('\n')
		
		if err != nil {
			log.Printf("unable to read, got an error: %v\n", err)
			continue
		}
		conn.Write([]byte(strIn))


	}


}
