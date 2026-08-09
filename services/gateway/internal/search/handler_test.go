package search_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsflow/gateway/internal/search"
)

func TestSearchHandler(t *testing.T) {
	// Mock Incident Service
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "payment" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"incidents":[{"id":"inc-1","title":"Payment DB latency"}]}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"incidents":[]}}`))
	}))
	defer incidentServer.Close()

	// Mock Registry Service
	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "payment" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"services":[{"id":"svc-1","name":"payment-service"}]}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"services":[]}}`))
	}))
	defer registryServer.Close()

	// Parse ports from server URLs
	incPort := incidentServer.URL[17:]
	regPort := registryServer.URL[17:]

	handler := search.NewSearchHandler("8081", incPort, regPort)

	// 1. Query without 'q' parameter -> 400 Bad Request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	w := httptest.NewRecorder()
	handler.Search(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing query, got %d", w.Code)
	}

	// 2. Query with q=payment -> 200 OK with aggregated results
	req = httptest.NewRequest(http.MethodGet, "/api/v1/search?q=payment", nil)
	w = httptest.NewRecorder()
	handler.Search(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp struct {
		Data search.SearchResult `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Data.Query != "payment" {
		t.Errorf("expected query payment, got %s", resp.Data.Query)
	}
	if len(resp.Data.Incidents) != 1 {
		t.Errorf("expected 1 incident, got %d", len(resp.Data.Incidents))
	}
	if len(resp.Data.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(resp.Data.Services))
	}
}
