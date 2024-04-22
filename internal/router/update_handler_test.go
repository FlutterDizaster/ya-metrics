package router

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

func TestRouter_updateHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		metric      view.Metric
	}
	type test struct {
		name    string
		request string
		method  string
		want    want
	}

	tests := []test{
		{
			name:    "simple test",
			request: "/update/counter/test/55",
			method:  http.MethodPost,
			want: want{
				code:        200,
				contentType: "text/plain charset=utf-8",
				metric: view.Metric{
					Kind:  "counter",
					Name:  "test",
					Value: "55",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MockMetricsStorage{
				content: make([]view.Metric, 0),
			}
			r := NewRouter(&Settings{
				Storage: storage,
			})

			server := httptest.NewServer(r)
			defer server.Close()

			client := resty.New()

			resp, err := client.R().Post(fmt.Sprintf("%s%s", server.URL, tt.request))

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.want.code, resp.StatusCode())

			assert.Contains(t, storage.content, tt.want.metric)
		})
	}
}
