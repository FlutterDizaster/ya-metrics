package handlers

import (
	"net/http"
	"testing"
)

func TestGetMetricHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		storage GetMetricStorage
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := GetMetricHandler{
				storage: tt.fields.storage,
			}
			h.ServeHTTP(tt.args.w, tt.args.r)
		})
	}
}
