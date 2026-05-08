package enums

type ParseState string

const (
	ParseStateDone    ParseState = "done"
	ParseStateInit    ParseState = "initialized"
	ParseStateHeaders ParseState = "headers"
	ParseStateBody    ParseState = "body"
)

type StatusCode string

const (
	StatusCodeOK                  StatusCode = "200"
	StatusCodeBadRequest          StatusCode = "400"
	StatusCodeInternalServerError StatusCode = "500"
)

type ContentType string

const (
	ContentTypeHTML  ContentType = "text/html"
	ContentTypePlain ContentType = "text/plain"
	ContentTypeVideo ContentType = "video/mp4"
)

func (c ContentType) String() string {
	return string(c)
}

type Network string

const (
	NetworkTCP Network = "tcp"
	NetworkUDP Network = "udp"
)

func (n Network) String() string {
	return string(n)
}
