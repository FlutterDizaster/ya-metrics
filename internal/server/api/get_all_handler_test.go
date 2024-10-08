package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_getAllHandler(t *testing.T) {
	type want struct {
		code    int
		content string
	}
	type test struct {
		name   string
		values []view.Metric
		want   want
	}
	tests := []test{
		{
			name: "simple test",
			values: []view.Metric{
				{
					ID:    "test",
					MType: "gauge",
					Value: func(i float64) *float64 { return &i }(555),
				},
			},
			want: want{
				code: 200,
				content: `
	<!doctype html>
	<html lang="en">
		<head>
			<title>Metrics</title>
		</head>
		<body>
			<table>
				<th>Kind</th>
				<th>Name</th>
				<th>Value</th>
					<tr>
						<td>gauge</td>
						<td>test</td>
						<td>555</td>
					</tr>
			</table>
		</body>
	</html>`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(&Settings{
				Storage: &MockMetricsStorage{
					content: tt.values,
				},
			})

			server := httptest.NewServer(http.HandlerFunc(r.getAllHandler))
			defer server.Close()

			client := resty.New()

			resp, err := client.R().Get(fmt.Sprintf("%s/", server.URL))

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.want.code, resp.StatusCode())

			body := string(resp.Body())
			body = strings.ReplaceAll(body, "\n", "")
			body = strings.ReplaceAll(body, "\t", "")

			testBoyd := strings.ReplaceAll(tt.want.content, "\n", "")
			testBoyd = strings.ReplaceAll(testBoyd, "\t", "")

			assert.Equal(t, testBoyd, body)
		})
	}
}

func TestAPI_getAllJSONHandler(t *testing.T) {
	type want struct {
		code    int
		content string
	}
	type test struct {
		name   string
		values view.Metrics
		want   want
	}
	tests := []test{
		{
			name: "simple test",
			values: []view.Metric{
				{
					ID:    "test",
					MType: "gauge",
					Value: func(i float64) *float64 { return &i }(555),
				},
			},
			want: want{
				code: 200,
			},
		},
	}

	for i := range tests {
		strContent, err := tests[i].values.MarshalJSON()
		require.NoError(t, err)
		tests[i].want.content = string(strContent)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(&Settings{
				Storage: &MockMetricsStorage{
					content: tt.values,
				},
			})

			server := httptest.NewServer(http.HandlerFunc(r.getAllJSONHandler))
			defer server.Close()

			client := resty.New()

			resp, err := client.R().Get(fmt.Sprintf("%s/", server.URL))

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.want.code, resp.StatusCode())

			var respMetrics view.Metrics
			err = respMetrics.UnmarshalJSON(resp.Body())
			require.NoError(t, err)

			assert.Equal(t, tt.values, respMetrics)
		})
	}
}
