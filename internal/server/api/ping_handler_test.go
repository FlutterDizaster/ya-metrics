package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI_pingHandler(t *testing.T) {
	tests := []struct {
		name      string
		expectErr bool
		code      int
	}{
		{
			name:      "no err",
			expectErr: false,
			code:      200,
		},
		{
			name:      "err",
			expectErr: true,
			code:      500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Settings{}

			if tt.expectErr {
				s.Storage = &MockMetricsStorage{pingErr: errors.New("err")}
			} else {
				s.Storage = &MockMetricsStorage{pingErr: nil}
			}
			r := New(s)

			server := httptest.NewServer(http.HandlerFunc(r.pingHandler))

			client := resty.New()

			resp, err := client.R().Get(fmt.Sprintf("%s/", server.URL))

			require.NoError(t, err, "error making http request")
			assert.Equal(t, tt.code, resp.StatusCode())
			server.Close()
		})
	}
}
