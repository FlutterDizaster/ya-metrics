package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (r *Router) getMetricHandler(w http.ResponseWriter, req *http.Request) {
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")

	value, err := r.storage.GetMetricValue(kind, name)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		http.Error(w, "", http.StatusTeapot)
	}
}
