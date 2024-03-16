package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testHandler struct {
}

func (th testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNewRouter(t *testing.T) {
	rs := RouterSettings{
		UpdateHandler:    testHandler{},
		GetAllHandler:    testHandler{},
		GetMetricHandler: testHandler{},
	}

	ts := httptest.NewServer(NewRouter(rs))
	defer ts.Close()

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
			req, err := http.NewRequest(tt.method, ts.URL+tt.path, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want, resp.StatusCode)
		})
	}
}
