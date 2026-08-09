package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opsflow/common/httputil"
)

type SearchHandler struct {
	authPort     string
	incidentPort string
	registryPort string
	client       *http.Client
}

func NewSearchHandler(authPort, incidentPort, registryPort string) *SearchHandler {
	return &SearchHandler{
		authPort:     authPort,
		incidentPort: incidentPort,
		registryPort: registryPort,
		client:       &http.Client{Timeout: 5 * time.Second},
	}
}

type SearchResult struct {
	Query     string        `json:"query"`
	Incidents []interface{} `json:"incidents"`
	Services  []interface{} `json:"services"`
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", httputil.RequestID(r.Context()))
		return
	}

	q := r.URL.Query().Get("q")
	if strings.TrimSpace(q) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Query parameter 'q' is required", httputil.RequestID(r.Context()))
		return
	}

	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "all"
	}

	res := SearchResult{
		Query:     q,
		Incidents: []interface{}{},
		Services:  []interface{}{},
	}

	// Fetch Incidents if type is all or incidents
	if searchType == "all" || searchType == "incidents" {
		incidents, _ := h.searchIncidents(r, q)
		res.Incidents = incidents
	}

	// Fetch Services if type is all or services
	if searchType == "all" || searchType == "services" {
		services, _ := h.searchServices(r, q)
		res.Services = services
	}

	httputil.WriteSuccess(w, http.StatusOK, res, httputil.RequestID(r.Context()))
}

func (h *SearchHandler) searchIncidents(r *http.Request, query string) ([]interface{}, error) {
	url := fmt.Sprintf("http://localhost:%s/api/v1/incidents?q=%s", h.incidentPort, query)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		return []interface{}{}, err
	}

	// Forward auth and correlation headers
	req.Header.Set("Authorization", r.Header.Get("Authorization"))
	req.Header.Set("X-Request-ID", httputil.RequestID(r.Context()))

	resp, err := h.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return []interface{}{}, nil
	}
	defer resp.Body.Close()

	var body struct {
		Data struct {
			Incidents []interface{} `json:"incidents"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return body.Data.Incidents, nil
}

func (h *SearchHandler) searchServices(r *http.Request, query string) ([]interface{}, error) {
	url := fmt.Sprintf("http://localhost:%s/api/v1/services?q=%s", h.registryPort, query)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		return []interface{}{}, err
	}

	req.Header.Set("Authorization", r.Header.Get("Authorization"))
	req.Header.Set("X-Request-ID", httputil.RequestID(r.Context()))

	resp, err := h.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return []interface{}{}, nil
	}
	defer resp.Body.Close()

	var body struct {
		Data struct {
			Services []interface{} `json:"services"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return body.Data.Services, nil
}
