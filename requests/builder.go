package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	errors  []error
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

// Post sets the request method to POST.
func (b *Builder) Post() *Builder {
	b.method = http.MethodPost

	return b
}

// Put sets the request method to PUT.
func (b *Builder) Put() *Builder {
	b.method = http.MethodPut

	return b
}

// Patch sets the request method to PATCH.
func (b *Builder) Patch() *Builder {
	b.method = http.MethodPatch

	return b
}

// Body sets the request body.
func (b *Builder) Body(body io.Reader) *Builder {
	b.body = body

	return b
}

// JSON sets the request body as json.
func (b *Builder) JSON(v any) *Builder {
	bs, err := json.Marshal(v)
	if err != nil {
		b.errors = append(b.errors, err)
		return b
	}

	b.body = bytes.NewReader(bs)
	b.headers = append(b.headers, headerValue{"Content-Type", "application/json"})
	b.headers = append(b.headers, headerValue{"Accept", "application/json"})
	return b
}

// Header adds a request header.
func (b *Builder) Header(key, value string) *Builder {
	b.headers = append(b.headers, headerValue{key, value})

	return b
}

// Headers adds all request headers inside values map input.
func (b *Builder) Headers(values map[string]string) *Builder {
	for k, v := range values {
		b.headers = append(b.headers, headerValue{k, v})
	}

	return b
}

// Build builds the Request.
func (b *Builder) Build(ctx context.Context) (*http.Request, error) {
	if len(b.errors) != 0 {
		return nil, errors.Join(b.errors...)
	}

	req, err := http.NewRequestWithContext(ctx, b.method, b.url, b.body)
	if err != nil {
		return nil, err
	}

	for _, hv := range b.headers {
		req.Header.Add(hv.k, hv.v)
	}

	return req, nil
}
