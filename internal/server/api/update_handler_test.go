package api

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/FlutterDizaster/ya-metrics/internal/view"
// 	"github.com/go-resty/resty/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestRouter_updateHandler(t *testing.T) {
// 	type want struct {
// 		code        int
// 		contentType string
// 		metric      view.Metric
// 	}
// 	type test struct {
// 		name    string
// 		request string
// 		method  string
// 		want    want
// 	}

// 	tests := []test{
// 		{
// 			name:    "simple test",
// 			request: "/update/counter/test/55",
// 			method:  http.MethodPost,
// 			want: want{
// 				code:        200,
// 				contentType: "text/plain charset=utf-8",
// 				metric: view.Metric{
// 					MType: counter,
// 					ID:    "test",
// 					Delta: func(i int64) *int64 { return &i }(55),
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			storage := &MockMetricsStorage{
// 				content: make([]view.Metric, 0),
// 			}
// 			r := NewRouter(&Settings{
// 				Storage: storage,
// 			})

// 			server := httptest.NewServer(r)
// 			defer server.Close()

// 			client := resty.New()

// 			resp, err := client.R().Post(fmt.Sprintf("%s%s", server.URL, tt.request))

// 			require.NoError(t, err, "error making http request")
// 			assert.Equal(t, tt.want.code, resp.StatusCode())

// 			assert.Contains(t, storage.content, tt.want.metric)
// 		})
// 	}
// }

// func TestRouter_updateJSONHandler(t *testing.T) {
// 	type want struct {
// 		code        int
// 		contentType string
// 	}
// 	type test struct {
// 		name    string
// 		request view.Metric
// 		method  string
// 		want    want
// 	}

// 	tests := []test{
// 		{
// 			name: "simple test",
// 			request: func() view.Metric {
// 				metric, _ := view.NewMetric("counter", "PollCounter", "55")
// 				return *metric
// 			}(),
// 			method: http.MethodPost,
// 			want: want{
// 				code:        200,
// 				contentType: "application/json charset=utf-8",
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			storage := &MockMetricsStorage{
// 				content: make([]view.Metric, 0),
// 			}
// 			r := NewRouter(&Settings{
// 				Storage: storage,
// 			})

// 			server := httptest.NewServer(r)
// 			defer server.Close()

// 			client := resty.New()

// 			resp, err := client.R().
// 				SetBody(tt.request).
// 				Post(fmt.Sprintf("%s%s", server.URL, "/update"))

// 			require.NoError(t, err, "error making http request")

// 			var respMetric view.Metric
// 			err = json.Unmarshal(resp.Body(), &respMetric)

// 			require.NoError(t, err, "error unmarshaling metric")

// 			assert.Equal(t, respMetric, tt.request)
// 			assert.Equal(t, tt.want.code, resp.StatusCode())

// 			assert.Contains(t, storage.content, tt.request)
// 		})
// 	}
// }
