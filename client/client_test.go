package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDoer struct {
	res      *http.Response
	err      error
	attempts uint8
}

func (td *testDoer) Do(req *http.Request) (*http.Response, error) {
	if req.Context().Err() != nil {
		return nil, req.Context().Err()
	}

	td.attempts++
	return td.res, td.err
}

func response(code int) *http.Response {
	return &http.Response{
		StatusCode: code,
		Status:     http.StatusText(code),
	}
}

func errHTTPClientTimeout() error {
	return fmt.Errorf("%w (Client.Timeout exceeded while awaiting headers)", context.DeadlineExceeded)
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name     string
		doer     *testDoer
		res      *http.Response
		attempts uint8
		err      string
	}{
		{name: "ok", doer: &testDoer{res: response(200)}, res: response(200), attempts: 1},
		{name: "error", doer: &testDoer{err: io.ErrUnexpectedEOF}, err: "unexpected EOF", attempts: 4},
		{name: "context cancelled", doer: &testDoer{err: context.Canceled}, err: "context canceled", attempts: 1},
		{name: "context deadline exceeded", doer: &testDoer{err: context.DeadlineExceeded}, err: "context deadline exceeded", attempts: 1},
		{name: "http client timeout", doer: &testDoer{err: errHTTPClientTimeout()}, err: "context deadline exceeded (Client.Timeout exceeded while awaiting headers)", attempts: 4},
		{name: "408", doer: &testDoer{res: response(408)}, res: response(408), attempts: 4},
		{name: "429", doer: &testDoer{res: response(429)}, res: response(429), attempts: 1},
		{name: "500", doer: &testDoer{res: response(500)}, res: response(500), attempts: 4},
		{name: "502", doer: &testDoer{res: response(502)}, res: response(502), attempts: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(Timeout(time.Minute), Retries(3))
			client.doer = tt.doer
			client.waitTime = func(uint8) time.Duration {
				return 0
			}

			res, err := client.Do(&http.Request{})
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.attempts, tt.doer.attempts)
			assert.Equal(t, tt.res, res)
		})
	}
}

func Test_defaultWaitTime(t *testing.T) {
	client := New()
	client.waitTime = defaultWaitTime(rand.New(rand.NewSource(1)))

	tests := []struct {
		i    uint8
		want string
	}{
		{0, "81ms"},
		{1, "1.887s"},
		{2, "4.847s"},
		{3, "9.059s"},
		{4, "16.081s"},
		{5, "25.318s"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("i=%d", tt.i), func(t *testing.T) {
			assert.Equal(t, tt.want, client.waitTime(tt.i).String())
		})
	}
}

func TestClient_AvoidSleepOnLastAttempt(t *testing.T) {
	timeFunc := func(i uint8) time.Duration {
		panic("unexpected")
	}

	client := New(Retries(0))
	client.doer = &testDoer{err: io.ErrUnexpectedEOF}
	client.waitTime = timeFunc

	res, err := client.Do(&http.Request{})
	require.EqualError(t, err, "unexpected EOF")
	assert.Nil(t, res)
}

func TestClient_context(t *testing.T) {
	t.Run("timeout", func(t *testing.T) {
		client := New()
		client.doer = &testDoer{err: io.ErrUnexpectedEOF}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, "", nil)
		require.NoError(t, err)

		res, err := client.Do(req)
		require.EqualError(t, err, "context deadline exceeded")
		assert.Nil(t, res)
	})
	t.Run("cancel", func(t *testing.T) {
		client := New()
		client.doer = &testDoer{err: io.ErrUnexpectedEOF}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, "", nil)
		require.NoError(t, err)

		res, err := client.Do(req)
		require.EqualError(t, err, "context canceled")
		assert.Nil(t, res)
	})
}
