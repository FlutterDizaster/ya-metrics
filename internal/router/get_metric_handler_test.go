package router

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_getMetricHandler(t *testing.T) {
	type want struct {
		code    int
		content string
	}
	type test struct {
		name   string
		values []view.Metric
		reqURL string
		want   want
	}
	tests := []test{
		{
			name: "simple test",
			values: []view.Metric{
				{
					Name:  "test1",
					Kind:  "gauge",
					Value: "54",
				},
			},
			reqURL: "/value/gauge/test1",
			want: want{
				code:    200,
				content: "54",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter(&Settings{
				Storage: &MockMetricsStorage{
					content: tt.values,
				},
			})

			server := httptest.NewServer(r)
			defer server.Close()

			client := resty.New()

			resp, err := client.R().Get(fmt.Sprintf("%s%s", server.URL, tt.reqURL))

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.want.code, resp.StatusCode())

			body := string(resp.Body())

			assert.Equal(t, tt.want.content, body)
		})
	}
}