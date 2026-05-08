# 🌐 HTTP Server from Scratch in Go

> A ground-up implementation of an HTTP/1.1 server built in Golang — no frameworks, no magic, just raw TCP and the protocol itself.

---

## 📖 About

This project is a full implementation of an HTTP/1.1 server written in Go from scratch. The goal is to deeply understand how the web works under the hood — from raw TCP connections all the way up to chunked encoding and binary data handling.

---

## 🚀 Features

- Raw TCP listener and connection handling
- HTTP/1.1 request parsing (request line, headers, body)
- HTTP response construction and writing
- Chunked transfer encoding
- Binary data support
- Fully handwritten — zero external HTTP libraries

---

## 🏗️ Project Structure

```
.
├── main.go           # Entry point, starts the TCP listener
├── server/
│   ├── server.go     # TCP server setup and connection loop
│   ├── request.go    # HTTP request parsing
│   ├── response.go   # HTTP response building and writing
│   └── headers.go    # HTTP header parsing utilities
└── README.md
```

---

## 🛠️ Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.21+

### Run the Server

```bash
git clone https://github.com/Suryansh0301/HTTP-From-TCP.git
cd HTTP-From-TCP
go run main.go
```

The server will start listening on `localhost:42069` by default.

### Test It

```bash
curl -v http://localhost:42069/
```

---
## 🧠 How It Works

### TCP → HTTP

The server opens a raw TCP socket and listens for connections. Each connection is read byte-by-byte to parse the HTTP request manually:

1. **Request line** — method, path, protocol version
2. **Headers** — key-value pairs terminated by `\r\n`
3. **Body** — read based on `Content-Length` or chunked encoding

Responses are constructed as raw strings following the HTTP/1.1 spec and written back to the TCP stream.

### Chunked Encoding

For chunked responses, each chunk is prefixed with its size in hexadecimal, followed by the data, and terminated with a `0`-length chunk.

```
HTTP/1.1 200 OK
Transfer-Encoding: chunked

7\r\n
Mozilla\r\n
0\r\n
\r\n
```

---
