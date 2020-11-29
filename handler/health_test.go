package handler

import (
	"dev.volix.ops/thor/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	ms := storage.NewSimpleMetricStorage()

	req, err := http.NewRequest("GET", "/-/healthy", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	Health(ms).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
