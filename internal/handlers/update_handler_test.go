package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type TestStorage struct {
}

func (s *TestStorage) AddMetricValue(kind string, name string, value string) error {
	return nil
}

func TestUpdateHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code        int
		contentType string
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
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			storage := TestStorage{}

			handler := NewUpdateHandler(&storage)
			srv := httptest.NewServer(handler)

			req := resty.New().R()
			req.Method = test.method
			req.URL = srv.URL + test.request

			resp, err := req.Send()

			assert.NoError(t, err, "error making http request")
			assert.Equal(t, test.want.code, resp.StatusCode())

			//Test is useless
			//TODO: Edit

			srv.Close()
		})
	}
}
