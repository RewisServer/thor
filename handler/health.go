package handler

import (
	"dev.volix.ops/thor/storage"
	"io"
	"net/http"
)

func Health(ms *storage.MetricStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		err := ms.Healthy()
		if err == nil {
			w.WriteHeader(200)
			_, _ = io.WriteString(w, "OK")
		} else {
			http.Error(w, err.Error(), 500)
		}
	}
}
