// Package server provides code for an HTTP Server
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/peter-howell/httpfromtcp/internal/request"
	"github.com/peter-howell/httpfromtcp/internal/response"
	"github.com/peter-howell/httpfromtcp/internal/server"
)



const port = 42069

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	var body = ""
	code := response.StatusOK

	switch target {
	case "/yourproblem":
		code = response.StatusBadRequest
		body = "<html>\n" +
			"  <head>\n" +
			"    <title>400 Bad Request</title>\n" +
			"  </head>\n" +
			"  <body>\n" +
			"    <h1>Bad Request</h1>\n" +
			"    <p>Your request honestly kinda sucked.</p>\n" +
			"  </body>\n" +
			"</html>\n"
	case "/myproblem":
		code = response.StatusInternalServerError
		body = "<html>\n" +
			"  <head>\n" +
			"    <title>500 Internal Server Error</title>\n" +
			"  </head>\n" +
			"  <body>\n" +
			"    <h1>Internal Server Error</h1>\n" +
			"    <p>Okay, you know what? This one is on me.</p>\n" +
			"  </body>\n" +
			"</html>\n"
	default:
		body = "<html>\n" +
			"  <head>\n" +
			"    <title>200 OK</title>\n" +
			"  </head>\n" +
			"  <body>\n" +
			"    <h1>Success!</h1>\n" +
			"    <p>Your request was an absolute banger.</p>\n" +
			"  </body>\n" +
			"</html>\n"
	}
	err := w.WriteStatusLine(code)
	if err != nil {
		fmt.Printf("Got an error\n%v\n", err)
	}
	headers := response.GetDefaultHeaders(len(body))
	headers.Replace("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody([]byte(body))
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}






