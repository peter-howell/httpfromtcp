// Package server provides code for an HTTP Server
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/peter-howell/httpfromtcp/internal/encoding"
	"github.com/peter-howell/httpfromtcp/internal/headers"
	"github.com/peter-howell/httpfromtcp/internal/request"
	"github.com/peter-howell/httpfromtcp/internal/response"
	"github.com/peter-howell/httpfromtcp/internal/server"
)


func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	if strings.HasPrefix(target, "/httpbin/") {
		handleProxy(w, req)
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
	case "/video":
		handleVideo(w, req)
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

func handleProxy(w *response.Writer, req *request.Request) {
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
	h.Set("Trailers", "X-Content-SHA256")
	h.Set("Trailers", "X-Content-Length")
	w.WriteHeaders(h)

	bodyLen := 0
	currentChunkSize := 0
	const maxChunkSize = 1024
	currentChunkBuf := make([]byte, maxChunkSize)
	totalBodyBuf := make([]byte, 2048)

	for {
		currentChunkSize, err = resp.Body.Read(currentChunkBuf)

		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("Got an error reading the body, %v\n", err)
			break
		}

		if currentChunkSize < 1 {
			break
		}

		fmt.Println("Read", currentChunkSize, "bytes from response body")

		_, err = w.WriteChunkedBody(currentChunkBuf[:currentChunkSize])

		if err != nil {
			fmt.Printf("Got an error writing chunked body %v\n", err)
			break
		}

		if currentCapacity := len(totalBodyBuf); currentCapacity <= bodyLen + currentChunkSize {
			neededSize := int(math.Pow(2, math.Ceil(math.Log2(float64(currentCapacity + currentChunkSize)))))
			tem := make([]byte, neededSize)
			copy(tem, totalBodyBuf)
			copy(tem[bodyLen:], currentChunkBuf[:currentChunkSize])
			totalBodyBuf = tem
		} else {
			copy(totalBodyBuf[bodyLen:], currentChunkBuf[:currentChunkSize])
		}
		bodyLen += currentChunkSize

	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Printf("Got an error writing to done body, %v\n", err)
	}
	trailers := headers.NewHeaders()

	hash := encoding.SHA256Sum(totalBodyBuf[:bodyLen])
	hashStr := fmt.Sprintf("%x", hash)
	trailers.Set("X-Content-SHA256", hashStr)
	fmt.Println("hash: ", hashStr)
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", bodyLen))

	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Printf("got an error when writing trailers: %v\n", err)
	}
	
}

func handleVideo(w *response.Writer, req *request.Request) {
	
	
	w.WriteStatusLine(response.StatusOK)
	h := headers.NewHeaders()
	h.Set("Content-Type", "video/mp4")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Connection", "close")
	h.Set("Trailers", "X-Content-SHA256")
	h.Set("Trailers", "X-Content-Length")
	w.WriteHeaders(h)

	fname := "assets/vim.mp4"

	file, err := os.Open(fname)

	if err != nil {
		handle500(w, req)
		return
	}

	defer file.Close()

	bodyLen := 0
	currentChunkSize := 0
	const maxChunkSize = 1024
	currentChunkBuf := make([]byte, maxChunkSize)
	totalBodyBuf := make([]byte, 2048)

	for {
		currentChunkSize, err = file.Read(currentChunkBuf)

		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("Got an error reading the body, %v\n", err)
			break
		}

		if currentChunkSize < 1 {
			fmt.Printf("Got chunk size %d, and breaking the loop\n", currentChunkSize)
			break
		}

		fmt.Println("Read", currentChunkSize, "bytes from response body")

		_, err = w.WriteChunkedBody(currentChunkBuf[:currentChunkSize])

		if err != nil {
			fmt.Printf("Got an error writing chunked body %v\n", err)
			break
		}

		if currentCapacity := len(totalBodyBuf); currentCapacity <= bodyLen + currentChunkSize {
			neededSize := int(math.Pow(2, math.Ceil(math.Log2(float64(currentCapacity + currentChunkSize)))))
			tem := make([]byte, neededSize)
			copy(tem, totalBodyBuf)
			copy(tem[bodyLen:], currentChunkBuf[:currentChunkSize])
			totalBodyBuf = tem
		} else {
			copy(totalBodyBuf[bodyLen:], currentChunkBuf[:currentChunkSize])
		}
		bodyLen += currentChunkSize

	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Printf("Got an error writing to done body, %v\n", err)
	}
	trailers := headers.NewHeaders()

	hash := encoding.SHA256Sum(totalBodyBuf[:bodyLen])
	hashStr := fmt.Sprintf("%x", hash)
	trailers.Set("X-Content-SHA256", hashStr)
	fmt.Println("hash: ", hashStr)
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", bodyLen))

	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Printf("got an error when writing trailers: %v\n", err)
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

