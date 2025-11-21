// Package server provides code for an HTTP Server
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/peter-howell/httpfromtcp/internal/headers"
	"github.com/peter-howell/httpfromtcp/internal/request"
	"github.com/peter-howell/httpfromtcp/internal/response"
	"github.com/peter-howell/httpfromtcp/internal/server"
)


func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	if strings.HasPrefix(target, "/httpbin/") {
		handleChunked(w, req)
		return
	}
	var body = ""
	code := response.StatusOK

	switch target {
	case "/yourproblem":
		handle400(w, req)
		return
	case "/myproblem":
		handle500(w, req)
		return
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
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}

func handleChunked(w *response.Writer, req *request.Request) {
	url := fmt.Sprintf("https://httpbin.org/%s", strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/"))

	resp, err := http.Get(url)
	if err != nil {
		handle500(w, req)
		return
	}
	
	w.WriteStatusLine(response.StatusOK)
	h := headers.NewHeaders()
	h.Set("Content-Type", "text/plain")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Connection", "close")
	w.WriteHeaders(h)

	buf := make([]byte, 1024)

	for {
		n, err := resp.Body.Read(buf)

		if n <= 0 {
			break
		}

		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("Got an error reading the body, %v\n", err)
			break
		}

		_, err = w.WriteChunkedBody(buf[:n])

		if err != nil {
			fmt.Printf("Got an error writing chunked body %v\n", err)
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Printf("Got an error writing to done body, %v\n", err)
	}
}

func handle500(w *response.Writer, _req *request.Request) {

	code := response.StatusInternalServerError
	body := "<html>\n" +
		"  <head>\n" +
		"    <title>500 Internal Server Error</title>\n" +
		"  </head>\n" +
		"  <body>\n" +
		"    <h1>Internal Server Error</h1>\n" +
		"    <p>Okay, you know what? This one is on me.</p>\n" +
		"  </body>\n" +
		"</html>\n"

	err := w.WriteStatusLine(code)
	if err != nil {
		fmt.Printf("Got an error\n%v\n", err)
	}
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}

func handle400(w *response.Writer, _req *request.Request) {

	code := response.StatusBadRequest
	body := "<html>\n" +
		"  <head>\n" +
		"    <title>400 Bad Request</title>\n" +
		"  </head>\n" +
		"  <body>\n" +
		"    <h1>Bad Request</h1>\n" +
		"    <p>Your request honestly kinda sucked.</p>\n" +
		"  </body>\n" +
		"</html>\n"

	err := w.WriteStatusLine(code)
	if err != nil {
		fmt.Printf("Got an error\n%v\n", err)
	}
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody([]byte(body))
}


const port = 42069

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

