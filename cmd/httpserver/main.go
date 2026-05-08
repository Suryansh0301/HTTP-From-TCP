package main

import (
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"http-from-tcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

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

func handler(w *response.Writer, req *request.Request) {
	var statusCode response.StatusCode
	var responseByte []byte
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		statusCode = response.StatusCodeBadRequest
		responseByte = []byte(`
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
	case "/myproblem":
		statusCode = response.StatusCodeInternalServerError
		responseByte = []byte(`
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
	default:
		statusCode = response.StatusCodeOK
		responseByte = []byte(`<html>
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
