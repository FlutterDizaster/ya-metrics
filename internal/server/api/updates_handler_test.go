package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI_updateBatchHandler(t *testing.T) {
	tests := []struct {
		name   string
		values view.Metrics
		code   int
		dbErr  bool
	}{
		{
			name: "simple test",
			values: view.Metrics{
				{
					ID:    view.KindCounter,
					MType: view.KindCounter,
					Delta: func(i int64) *int64 { return &i }(45),
				},
				{
					ID:    view.KindGauge,
					MType: view.KindGauge,
					Value: func(i float64) *float64 { return &i }(54),
				},
			},
			code: 200,
		},
		{
			name: "db error test",
			values: view.Metrics{
				{
					ID:    view.KindCounter,
					MType: view.KindCounter,
					Delta: func(i int64) *int64 { return &i }(45),
				},
			},
			code:  400,
			dbErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MockMetricsStorage{}

			if tt.dbErr {
				storage.err = errors.New("err")
			}

			r := New(&Settings{
				Storage: storage,
			})

			server := httptest.NewServer(http.HandlerFunc(r.updateBatchHandler))

			client := resty.New()

			reqBody, err := tt.values.MarshalJSON()
			require.NoError(t, err)

			req := client.R().SetBody(reqBody)
			resp, err := req.Post(fmt.Sprintf("%s/", server.URL))

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.code, resp.StatusCode())

			if !tt.dbErr {
				assert.Equal(t, tt.values, storage.content)
			}
			server.Close()
		})
	}
}
