package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/handlers"
	"github.com/FlutterDizaster/ya-metrics/internal/view"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testGetAllStorage struct {
	content []view.Metric
}

func (s *testGetAllStorage) ReadAllMetrics() []view.Metric {
	return s.content
}

func TestGetAllHandler_ServeHTTP(t *testing.T) {
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
					Name:  "test",
					Kind:  "gauge",
					Value: "555",
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
			storage := testGetAllStorage{
				content: tt.values,
			}

			handler := handlers.NewGetAllHandler(&storage)
			srv := httptest.NewServer(handler)

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL

			resp, err := req.Send()

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.want.code, resp.StatusCode())

			body := string(resp.Body())
			body = strings.ReplaceAll(body, "\n", "")
			body = strings.ReplaceAll(body, "\t", "")

			testBoyd := strings.ReplaceAll(tt.want.content, "\n", "")
			testBoyd = strings.ReplaceAll(testBoyd, "\t", "")

			assert.Equal(t, testBoyd, body)

			srv.Close()
		})
	}
}
