package appinfoprinter

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintAppInfo(t *testing.T) {
	type test struct {
		name       string
		info       AppInfo
		wantOutput string
	}
	tests := []test{
		{
			name: "simple test",
			info: AppInfo{
				Version: "1.0.0",
				Date:    "2020-01-01",
				Commit:  "123456",
			},
			wantOutput: "Build version: 1.0.0\nBuild date: 2020-01-01\nBuild commit: 123456\n",
		},
		{
			name:       "N\\A test",
			info:       AppInfo{},
			wantOutput: "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStdout := os.Stdout

			reader, writer, err := os.Pipe()
			require.NoError(t, err)

			os.Stdout = writer

			err = PrintAppInfo(tt.info)
			require.NoError(t, err)

			writer.Close()

			var buf bytes.Buffer
			_, err = io.Copy(&buf, reader)
			require.NoError(t, err)

			reader.Close()

			require.Equal(t, tt.wantOutput, buf.String())

			os.Stdout = originalStdout
		})
	}
}
