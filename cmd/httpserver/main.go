package main

import (
	"crypto/sha256"
	"encoding/hex"
	"http-from-tcp/constants"
	"http-from-tcp/enums"
	headersPkg "http-from-tcp/internal/headers"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"http-from-tcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

// main initializes and starts the custom TCP-based HTTP server.
// The server is used to test different HTTP response behaviors such as:
//
//   - status line writing
//   - headers and trailers
//   - chunked transfer encoding
//   - binary file responses
//   - custom response body handling
//
// The server shuts down gracefully on SIGINT or SIGTERM.
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

// handler is the main request router for the custom HTTP server.
//
// Different routes are intentionally designed to test specific
// HTTP server features and response-writing methods.
//
// Routes:
//
//	/yourproblem -> tests 400 Bad Request responses
//	/myproblem   -> tests 500 Internal Server Error responses
//	/httpbin/*   -> tests chunked transfer encoding and trailers
//	/video       -> tests binary file streaming and content-type handling
//	default      -> tests standard 200 OK HTML responses
func handler(w *response.Writer, req *request.Request) {

	var statusCode enums.StatusCode
	var responseByte []byte

	target := req.RequestLine.RequestTarget

	switch {
	case target == "/yourproblem":
		statusCode = enums.StatusCodeBadRequest
		responseByte = request400()
	case target == "/myproblem":
		statusCode = enums.StatusCodeInternalServerError
		responseByte = request500()
	case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"):
		handleHttpBin(w, target)
		return

	case target == "/video":
		{
			handleVideo(w)
			return
		}
	default:
		statusCode = enums.StatusCodeOK
		responseByte = request200()
	}

	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Println(err)
	}

	headers := w.GetDefaultHeaders(len(responseByte))
	headers.Replace(constants.ContentType, enums.ContentTypeHTML.String())
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Println(err)
	}

	_, err = w.WriteBody(responseByte)
	if err != nil {
		log.Println(err)
	}
}

func handleHttpBin(w *response.Writer, target string) {
	res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
	if err != nil {
		w.WriteStatusLine(enums.StatusCodeInternalServerError)

		headers := w.GetDefaultHeaders(len(request500()))
		w.WriteHeaders(headers)
		w.WriteBody(request500())
		return
	}
	defer res.Body.Close()

	w.WriteStatusLine(enums.StatusCodeOK)

	headers := w.GetDefaultHeaders(0)
	headers.Delete(constants.ContentLength)
	headers.Set(constants.TransferEncoding, "chunked")
	headers.Set(constants.Trailer, constants.XContentSHA256)
	headers.Set(constants.Trailer, constants.XContentLength)

	w.WriteHeaders(headers)

	fullBody := []byte{}

	for {
		data := make([]byte, 64)

		n, err := res.Body.Read(data)
		if err != nil {
			break
		}

		fullBody = append(fullBody, data[:n]...)
		w.WriteChunkedBody(data[:n])
	}

	w.WriteChunkedBodyDone()

	trailers := headersPkg.NewHeaders()

	out := sha256.Sum256(fullBody)

	trailers.Set(constants.XContentSHA256, hex.EncodeToString(out[:]))
	trailers.Set(constants.XContentLength, strconv.Itoa(len(fullBody)))

	w.WriteTrailers(trailers)
}

func handleVideo(w *response.Writer) {
	f, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Println("read file error:", err)
		return
	}

	headers := w.GetDefaultHeaders(len(f))
	headers.Replace(constants.ContentType, enums.ContentTypeVideo.String())

	w.WriteStatusLine(enums.StatusCodeOK)
	w.WriteHeaders(headers)
	w.WriteBody(f)
}
