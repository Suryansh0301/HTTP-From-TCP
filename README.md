# HTTP From TCP — HTTP/1.1 Server in Go

A ground-up implementation of an HTTP/1.1 server in Go. No `net/http`, no frameworks — just a raw TCP socket, the protocol specification, and the parser written by hand.

---

## Architecture

```
  TCP Socket (port 42069)
        │
        ▼
  ┌─────────────────────┐
  │     TCP Listener    │  accepts raw connections
  └──────────┬──────────┘
             │  goroutine per connection
             ▼
  ┌─────────────────────┐
  │   Request Parser    │  byte stream → Request struct
  │                     │
  │  1. Request line    │  method, path, HTTP version
  │  2. Headers         │  key-value, CRLF-terminated
  │  3. Body            │  Content-Length or chunked
  └──────────┬──────────┘
             │
             ▼
  ┌─────────────────────┐
  │   Request Handler   │  routes and processes request
  └──────────┬──────────┘
             │
             ▼
  ┌─────────────────────┐
  │   Response Writer   │  constructs and writes HTTP response
  │                     │
  │  - Status line      │
  │  - Headers          │
  │  - Body / Chunks    │
  └──────────┬──────────┘
             │
             ▼
       TCP Write
```

Each layer has a single responsibility and is independently testable.

---

## Protocol Implementation

### Request Parsing

HTTP over TCP is a byte stream with no inherent message boundaries. The parser reads incrementally and advances through three distinct states:

- Request line — method, path, and protocol version extracted from the first `\r\n`-terminated line
- Headers — key-value pairs parsed until the blank line (`\r\n\r\n`) signaling end of headers
- Body — read according to `Content-Length` or decoded from chunked transfer encoding

Parse state is tracked explicitly via enums, making partial reads and error states first-class concerns rather than edge cases.

### Chunked Transfer Encoding

Chunked encoding allows responses to be sent without knowing the total content length upfront. Each chunk is prefixed with its byte size in hexadecimal, followed by the data, terminated by a zero-length chunk.

```
HTTP/1.1 200 OK
Transfer-Encoding: chunked

7\r\n
Mozilla\r\n
9\r\n
Developer\r\n
0\r\n
\r\n
```

This is implemented in the response writer and supports streaming binary data — including video files.

### Binary Data

Binary responses (e.g. `vim.mp4`) are handled correctly by treating the body as raw bytes rather than strings, and setting the appropriate `Content-Type` header. No encoding assumptions are made about the payload.

---

## Project Structure

```
.
├── cmd/
│   ├── httpserver/
│   │   ├── assets/
│   │   │   └── vim.mp4          # Binary asset for testing binary response handling
│   │   └── main.go              # Entry point
│   ├── tcplistener/
│   │   └── main.go              # Raw TCP listener utility
│   └── udpsender/               # UDP sender utility
├── constants/
│   └── constants.go             # Shared constants (CRLF, header names)
├── enums/
│   └── enums.go                 # Status codes, content types, parse states
├── internal/
│   ├── headers/
│   │   ├── headers.go           # Header parsing
│   │   └── headers_test.go
│   ├── request/
│   │   ├── request.go           # Request parsing
│   │   └── request_test.go
│   └── response/
│       └── response.go          # Response construction and writing
├── server/
│   └── server.go                # TCP server and connection loop
├── go.mod
└── go.sum
```

---

## Getting Started

**Prerequisites:** Go 1.21+

```bash
git clone https://github.com/Suryansh0301/HTTP-From-TCP.git
cd HTTP-From-TCP
go run cmd/httpserver/main.go
```

The server listens on `localhost:42069` by default.

```bash
# Basic request
curl -v http://localhost:42069/

# Chunked response
curl -v http://localhost:42069/chunked

# Binary response
curl -v http://localhost:42069/video --output vim.mp4
```

---

## Testing

Unit tests cover header parsing and request parsing in isolation, without a live server.

```bash
go test ./...
```

The `internal/` structure keeps each component independently testable — headers, request parsing, and response writing are never coupled at the test level.

---

## References

- HTTP/1.1 Specification — https://www.rfc-editor.org/rfc/rfc9110
- Chunked Transfer Encoding — https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Transfer-Encoding
- OpenMyMind — Reading from TCP Streams — https://www.openmymind.net/2012/1/12/Reading-From-TCP-Streams/
- StackOverflow — What is a message boundary? — https://stackoverflow.com/questions/9563563/what-is-a-message-boundary
- Stephen Cleary — Message Framing — https://blog.stephencleary.com/2009/04/message-framing.html
- Beej's Guide to Network Programming — https://beej.us/guide/bgnet/html/
- Wikipedia — Transmission Control Protocol — https://en.wikipedia.org/wiki/Transmission_Control_Protocol
