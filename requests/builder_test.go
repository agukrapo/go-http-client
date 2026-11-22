package requests

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_builder(t *testing.T) {
	t.Run("method", func(t *testing.T) {
		req, err := New("_url").Method(http.MethodTrace).Build(t.Context())
		require.NoError(t, err)

		assert.Equal(t, http.MethodTrace, req.Method)
	})
	t.Run("post", func(t *testing.T) {
		req, err := New("_url").Post().Build(t.Context())
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, req.Method)
	})
	t.Run("put", func(t *testing.T) {
		req, err := New("_url").Put().Build(t.Context())
		require.NoError(t, err)

		assert.Equal(t, http.MethodPut, req.Method)
	})
	t.Run("patch", func(t *testing.T) {
		req, err := New("_url").Patch().Build(t.Context())
		require.NoError(t, err)

		assert.Equal(t, http.MethodPatch, req.Method)
	})
	t.Run("body", func(t *testing.T) {
		req, err := New("_url").Body(strings.NewReader("_body")).Build(t.Context())
		require.NoError(t, err)

		b, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Equal(t, "_body", string(b))
	})
	t.Run("header", func(t *testing.T) {
		req, err := New("_url").Header("_header_key", "_header_value").Build(t.Context())
		require.NoError(t, err)

		assert.Equal(t, "_header_value", req.Header.Get("_header_key"))
	})
	t.Run("headers", func(t *testing.T) {
		req, err := New("_url").Headers(map[string]string{"_header_key": "_header_value"}).Build(t.Context())
		require.NoError(t, err)

		assert.Equal(t, "_header_value", req.Header.Get("_header_key"))
	})
}

func TestBuilder_JSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		body := map[string]any{
			"_bool":   true,
			"_string": "qwerty",
			"_number": 3.14,
		}

		req, err := New("_url").JSON(body).Build(t.Context())
		require.NoError(t, err)

		b, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Equal(t, `{"_bool":true,"_number":3.14,"_string":"qwerty"}`, string(b))
		assert.Equal(t, `application/json`, req.Header.Get("Content-Type"))
		assert.Equal(t, `application/json`, req.Header.Get("Accept"))
	})
	t.Run("error", func(t *testing.T) {
		_, err := New("_url").JSON(context.AfterFunc).Build(t.Context())
		assert.EqualError(t, err, "json: unsupported type: func(context.Context, func()) func() bool")
	})
}
