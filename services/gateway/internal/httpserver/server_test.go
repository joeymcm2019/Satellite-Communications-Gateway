package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	r := httptest.NewRecorder()
	NewHandler().ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/health", nil))
	var body map[string]string
	if err := json.Unmarshal(r.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if r.Code != http.StatusOK || body["status"] != "ok" || r.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected response: %d %s", r.Code, r.Body.String())
	}
}
