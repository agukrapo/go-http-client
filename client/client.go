// Package client includes a retryable HTTP client and related types.
package client

import (
	"context"
	"errors"
	"iter"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client represents a retryable HTTP client.
type Client struct {
	doer     doer
	retries  uint8
	waitTime func(uint8) time.Duration
}

// Option represents Client constructor option.
type Option func(*Client)

// Retries represents an Option to specify Client retry attempts.
func Retries(retries uint8) Option {
	return func(c *Client) {
		c.retries = retries
	}
}

// Timeout represents an Option to specify Client timeout.
func Timeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.doer = &http.Client{
			Timeout: timeout,
		}
	}
}

// New creates a new Client.
func New(opts ...Option) *Client {
	out := &Client{
		doer: &http.Client{
			Timeout: 30 * time.Second,
		},
		retries:  6,
		waitTime: defaultWaitTime(rand.New(rand.NewSource(time.Now().Unix()))), // #nosec G404
	}

	for _, o := range opts {
		o(out)
	}

	return out
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(req *http.Request) (res *http.Response, err error) {
	for range attempts(req.Context(), c.retries, c.waitTime) {
		res, err = c.doer.Do(req)
		if !shouldRetry(res, err) {
			return res, err
		}
	}

	if req.Context().Err() != nil {
		err = req.Context().Err()
	}

	return res, err
}

func shouldRetry(res *http.Response, err error) bool {
	switch {
	case isHttpClientTimeout(err):
		return true
	case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
		return false
	case err != nil:
		return true
	case res == nil:
		return false
	case res.StatusCode == http.StatusRequestTimeout,
		strings.HasPrefix(strconv.Itoa(res.StatusCode), "5"):
		return true
	default:
		return false
	}
}

func isHttpClientTimeout(err error) bool {
	return err != nil && strings.HasSuffix(err.Error(), "(Client.Timeout exceeded while awaiting headers)")
}

// attempts encapsulates retry logic inside an iterator
func attempts(ctx context.Context, max uint8, wait func(uint8) time.Duration) iter.Seq[uint8] {
	var attempt uint8

	return func(yield func(uint8) bool) {
		if !yield(attempt) {
			return
		}

		for attempt < max {
			attempt++

			select {
			case <-time.After(wait(attempt)):
				if !yield(attempt) {
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}
}

func defaultWaitTime(r *rand.Rand) func(attempt uint8) time.Duration {
	return func(attempt uint8) time.Duration {
		backoff := time.Duration(math.Pow(float64(attempt), 2)) * time.Second
		jitter := time.Duration(r.Intn(1000)) * time.Millisecond
		return backoff + jitter
	}
}
