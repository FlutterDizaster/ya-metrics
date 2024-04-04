package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testGetValueStorage struct {
	content []struct {
		Name  string
		Kind  string
		Value string
	}
}

func (s *testGetValueStorage) GetMetricValue(_ string, name string) (string, error) {
	var value string
	var err error

	for _, v := range s.content {
		if v.Name == name {
			value = v.Value
		}
	}
	if value == "" {
		err = errors.New("Not found")
	}

	return value, err
}

func TestGetMetricHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code    int
		content string
	}
	type test struct {
		name   string
		values []struct {
			Name  string
			Kind  string
			Value string
		}
		reqURL string
		want   want
	}
	tests := []test{
		{
			name: "simple test",
			values: []struct {
				Name  string
				Kind  string
				Value string
			}{
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
			storage := testGetValueStorage{
				content: tt.values,
			}

			handler := handlers.NewGetMetricHandler(&storage)

			router := chi.NewRouter()
			router.Get("/value/{kind}/{name}", handler.ServeHTTP)

			srv := httptest.NewServer(router)

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL + tt.reqURL

			resp, err := req.Send()

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.want.code, resp.StatusCode())

			body := string(resp.Body())

			assert.Equal(t, tt.want.content, body)

			srv.Close()
		})
	}
}
