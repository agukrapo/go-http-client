// Package client includes a retryable HTTP client and related types.
package client

import (
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

// ResponseValidatorFunc returns an error when a http.Response should retry.
type ResponseValidatorFunc func(*http.Response) error

// WaitTimeFunc tells how much Client wait between failed attempts, starting from 0.
type WaitTimeFunc func(int) time.Duration

// Client represents a retryable HTTP client.
type Client struct {
	doer         doer
	waitTime     WaitTimeFunc
	resValidator ResponseValidatorFunc
	attempts     int
}

// Option represents Client constructor option.
type Option func(*Client)

// Attempts represents an Option to specify Client retry attempts.
func Attempts(attempts int) Option {
	return func(c *Client) {
		c.attempts = attempts
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

// WaitTime represents an Option to specify Client wait time between attempts.
func WaitTime(fn WaitTimeFunc) Option {
	return func(c *Client) {
		c.waitTime = fn
	}
}

// New creates a new Client.
func New(opts ...Option) *Client {
	out := &Client{
		doer: &http.Client{
			Timeout: 30 * time.Second,
		},
		resValidator: validate,
		waitTime:     defaultWaitTime(rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))),
		attempts:     6,
	}

	for _, o := range opts {
		o(out)
	}

	return out
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var res *http.Response
	err := c.retry(func() (err error) {
		res, err = c.doer.Do(req)
		if err != nil {
			return err
		}

		if err := c.resValidator(res); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) retry(f func() error) error {
	var err error

	for i := 0; i < c.attempts; i++ {
		if err = f(); err == nil {
			return nil
		}

		if i != c.attempts-1 { // avoid sleep on last attempt
			t := c.waitTime(i + 1)
			time.Sleep(t)
		}
	}

	return fmt.Errorf("after %d attempts: %w", c.attempts, err)
}

func defaultWaitTime(r *rand.Rand) WaitTimeFunc {
	return func(i int) time.Duration {
		v := time.Duration(math.Pow(float64(i), 2)) * (500 * time.Millisecond)
		r := time.Duration(r.IntN(1000)) * time.Millisecond
		return v + r
	}
}

func validate(res *http.Response) error {
	switch {
	case res == nil:
		return nil

	case res.StatusCode == http.StatusRequestTimeout,
		res.StatusCode == http.StatusTooManyRequests,
		strings.HasPrefix(strconv.Itoa(res.StatusCode), "5"):

		return fmt.Errorf("invalid status: %d %s", res.StatusCode, res.Status)
	}

	return nil
}
