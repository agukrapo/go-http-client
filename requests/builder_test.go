package requests

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_JSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		body := map[string]any{
			"_bool":   true,
			"_string": "qwerty",
			"_number": 3.14,
		}

		req, err := New("_url").JSON(body).Build(context.Background())
		require.NoError(t, err)

		b, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.Equal(t, `{"_bool":true,"_number":3.14,"_string":"qwerty"}`, string(b))
		assert.Equal(t, `application/json`, req.Header.Get("Content-Type"))
		assert.Equal(t, `application/json`, req.Header.Get("Accept"))
	})
	t.Run("error", func(t *testing.T) {
		_, err := New("_url").JSON(context.AfterFunc).Build(context.Background())
		assert.EqualError(t, err, "json: unsupported type: func(context.Context, func()) func() bool")
	})
}
