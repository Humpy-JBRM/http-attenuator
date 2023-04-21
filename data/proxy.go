package data

type ProxyRequest struct {
	// For tracing
	Id string `json:"id"`

	// The underlying HTTP request
	Request *HttpRequest `json:"http_request"`
}
