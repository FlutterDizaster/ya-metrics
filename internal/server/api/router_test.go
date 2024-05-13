package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter(&Settings{
		Storage: &MockMetricsStorage{
			content: make([]view.Metric, 0),
		},
	})

	server := httptest.NewServer(r)
	defer server.Close()

	tests := []struct {
		name   string
		path   string
		method string
		want   int
	}{
		{
			name:   "simple test",
			path:   "/update/gauge/test/55",
			method: http.MethodPost,
			want:   http.StatusOK,
		},
		{
			name:   "not found test",
			path:   "/update/gauge/555",
			method: http.MethodPost,
			want:   http.StatusNotFound,
		},
		{
			name:   "method not allowed test",
			path:   "/update/gauge/test/444",
			method: http.MethodGet,
			want:   http.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()

			resp, err := client.R().Execute(tt.method, fmt.Sprintf("%s%s", server.URL, tt.path))

			require.NoError(t, err)

			assert.Equal(t, tt.want, resp.StatusCode())
		})
	}
}
