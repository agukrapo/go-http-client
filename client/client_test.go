package client

import (
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
	res *http.Response
	err error
}

func (td *testDoer) Do(*http.Request) (*http.Response, error) {
	return td.res, td.err
}

func testWaitTime(int) time.Duration {
	return 0
}

func response(code int) *http.Response {
	return &http.Response{
		StatusCode: code,
		Status:     http.StatusText(code),
	}
}

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name string
		doer doer
		res  *http.Response
		err  string
	}{
		{name: "ok", doer: &testDoer{res: response(200)}, res: response(200)},
		{name: "error", doer: &testDoer{err: io.ErrUnexpectedEOF}, err: "after 6 attempts: unexpected EOF"},
		{name: "408", doer: &testDoer{res: response(408)}, err: "after 6 attempts: invalid status: 408 Request Timeout"},
		{name: "429", doer: &testDoer{res: response(429)}, err: "after 6 attempts: invalid status: 429 Too Many Requests"},
		{name: "500", doer: &testDoer{res: response(500)}, err: "after 6 attempts: invalid status: 500 Internal Server Error"},
		{name: "502", doer: &testDoer{res: response(502)}, err: "after 6 attempts: invalid status: 502 Bad Gateway"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(Timeout(time.Minute), WaitTime(testWaitTime))
			c.doer = tt.doer

			res, err := c.Do(&http.Request{})
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.res, res)
		})
	}
}

func Test_defaultWaitTime(t *testing.T) {
	rand.Seed(98723)
	tests := []struct {
		i    int
		want string
	}{
		{0, "1.653s"},
		{1, "2.845s"},
		{2, "4.973s"},
		{3, "8.385s"},
		{4, "16.484s"},
		{5, "32.413s"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("i=%d", tt.i), func(t *testing.T) {
			assert.Equal(t, tt.want, defaultWaitTime(tt.i).String())
		})
	}
}

func TestClient_AvoidSleepOnLastAttempt(t *testing.T) {
	wt := func(i int) time.Duration {
		if i == 2 {
			panic("unexpected")
		}
		return 0
	}

	c := New(Attempts(2), WaitTime(wt))
	c.doer = &testDoer{err: io.ErrUnexpectedEOF}

	res, err := c.Do(&http.Request{})
	require.EqualError(t, err, "after 2 attempts: unexpected EOF")
	assert.Nil(t, res)
}
