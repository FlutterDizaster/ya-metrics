package utils_test

import (
	"testing"

	"github.com/FlutterDizaster/ya-metrics/pkg/utils"

	"github.com/stretchr/testify/require"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedErr error
	}{
		{
			name:        "simple test",
			url:         "localhost:8080",
			expectedErr: nil,
		},
		{
			name:        "scheme test",
			url:         "http://192.168.0.1",
			expectedErr: utils.ErrNotEmptyScheme,
		},
		{
			name:        "host test",
			url:         ":8080",
			expectedErr: utils.ErrURLWithoutHost,
		},
		{
			name:        "invalid port 1",
			url:         "example.com:fff",
			expectedErr: utils.ErrInvalidPort,
		},
		{
			name:        "invalid port 2",
			url:         "example.com:99999999",
			expectedErr: utils.ErrInvalidPort,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidateURL(tt.url)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
