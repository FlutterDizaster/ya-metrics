package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		{
			name:    "404 test",
			request: "/test/test",
			method:  http.MethodPost,
			want: want{
				code:        404,
				contentType: "text/plain charset=utf-8",
			},
		},
		{
			name:    "405 test",
			request: "/test/test",
			method:  http.MethodGet,
			want: want{
				code:        405,
				contentType: "text/plain charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := TestStorage{}

			req := httptest.NewRequest(test.method, test.request, http.NoBody)

			w := httptest.NewRecorder()
			handler := NewUpdateHandler(&storage)

			handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}
