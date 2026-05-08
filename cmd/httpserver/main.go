package main

import (
	"fmt"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"http-from-tcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func request200() []byte {
	return []byte(`
	  <html>
		<head>
		<title>200 OK</title>
		</head>
		<body>
		<h1>Success!</h1>
		<p>Your request was an absolute banger.</p>
    </body>
    </html>
  `)
}

func request400() []byte {
	return []byte(`
		<html>
		<head>
		<title>400 Bad Request</title>
		</head>
		<body>
		<h1>Bad Request</h1>
		<p>Your request honestly kinda sucked.</p>
		</body>
		</html>
	`)
}

func request500() []byte {
	return []byte(`
		<html>
		<head>
		<title>500 Internal Server Error</title>
		</head>
		<body>
		<h1>Internal Server Error</h1>
		<p>Okay, you know what? This one is on me.</p>
		</body>
		</html>
`)
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

func handler(w *response.Writer, req *request.Request) {
	var statusCode response.StatusCode
	var responseByte []byte
	target := req.RequestLine.RequestTarget
	switch {
	case target == "/yourproblem":
		statusCode = response.StatusCodeBadRequest
		responseByte = request400()
	case target == "/myproblem":
		statusCode = response.StatusCodeInternalServerError
		responseByte = request500()
	case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"):
		res, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", target[len("/httpbin/"):]))
		if err != nil {
			statusCode = response.StatusCodeInternalServerError
			responseByte = request500()
		} else {
			w.WriteStatusLine(response.StatusCodeOK)
			headers := w.GetDefaultHeaders(0)
			headers.Delete("Content-Length")
			headers.Set("Transfer-Encoding", "chunked")

			for {
				data := make([]byte, 64)
				_, err := res.Body.Read(data)
				if err != nil {
					break
				}

				w.WriteChunkedBody(data)
			}
			w.WriteChunkedBodyDone()
			return
		}

	default:
		statusCode = response.StatusCodeOK
		responseByte = request200()
	}

	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Println(err)
	}

	headers := w.GetDefaultHeaders(len(responseByte))
	headers.Replace("Content-Type", "text/html")
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Println(err)
	}

	_, err = w.WriteBody(responseByte)
	if err != nil {
		log.Println(err)
	}
}
