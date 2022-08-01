package requests

import (
	"context"
	"io"
	"net/http"
)

type headerValue struct {
	k, v string
}

// Builder represents a request builder.
type Builder struct {
	url     string
	method  string
	body    io.Reader
	headers []headerValue
}

// New creates a new Builder.
func New(url string) *Builder {
	return &Builder{
		url:    url,
		method: http.MethodGet,
	}
}

// Method sets the request method (default get).
func (b *Builder) Method(method string) *Builder {
	b.method = method
	return b
}

// Body sets the request body.
func (b *Builder) Body(body io.Reader) *Builder {
	b.body = body
	return b
}

// Body add a request header.
func (b *Builder) Header(key, value string) *Builder {
	b.headers = append(b.headers, headerValue{key, value})
	return b
}

// Build builds the Request.
func (b *Builder) Build(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, b.method, b.url, b.body)
	if err != nil {
		return nil, err
	}

	for _, hv := range b.headers {
		req.Header.Add(hv.k, hv.v)
	}

	return req, nil
}
