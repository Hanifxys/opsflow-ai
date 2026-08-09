package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/opsflow/common/httputil"
	"github.com/opsflow/registry-service/internal/application"
	"github.com/opsflow/registry-service/internal/domain"
	"github.com/opsflow/registry-service/internal/ports"
)

type RegistryHandler struct {
	service *application.ServiceRegistryService
}

func NewRegistryHandler(service *application.ServiceRegistryService) *RegistryHandler {
	return &RegistryHandler{service: service}
}

type createServiceRequest struct {
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	OwnerTeam     string             `json:"owner_team"`
	Criticality   domain.Criticality `json:"criticality"`
	RepositoryURL string             `json:"repository_url"`
}

type addEnvironmentRequest struct {
	Environment    string `json:"environment"`
	BaseURL        string `json:"base_url"`
	HealthEndpoint string `json:"health_endpoint"`
}

type addDependencyRequest struct {
	DependsOnServiceID uuid.UUID `json:"depends_on_service_id"`
	DependencyType     string    `json:"dependency_type"`
	Critical           bool      `json:"critical"`
}

type addHealthCheckRequest struct {
	Name            string `json:"name"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	ExpectedStatus  int    `json:"expected_status"`
	TimeoutMS       int    `json:"timeout_ms"`
	IntervalSeconds int    `json:"interval_seconds"`
}

func (h *RegistryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.OwnerTeam == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "name and owner_team are required", httputil.RequestID(r.Context()))
		return
	}

	svc, err := h.service.CreateService(r.Context(), req.Name, req.Description, req.OwnerTeam, req.Criticality, req.RepositoryURL)
	if err != nil {
		if errors.Is(err, domain.ErrServiceNameExists) {
			httputil.WriteError(w, http.StatusConflict, "NAME_EXISTS", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create service", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, svc, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ports.ServiceFilter{
		OwnerTeam:   q.Get("owner_team"),
		Criticality: q.Get("criticality"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	services, total, err := h.service.ListServices(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list services", httputil.RequestID(r.Context()))
		return
	}

	resp := map[string]interface{}{
		"services": services,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	}

	httputil.WriteSuccess(w, http.StatusOK, resp, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) Get(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	svc, err := h.service.GetService(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrServiceNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Service not found", httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get service", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, svc, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) Delete(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.service.DeleteService(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete service", httputil.RequestID(r.Context()))
		return
	}
	httputil.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Service deleted"}, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) AddEnvironment(w http.ResponseWriter, r *http.Request, serviceID uuid.UUID) {
	var req addEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Environment == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "environment name is required", httputil.RequestID(r.Context()))
		return
	}

	env, err := h.service.AddEnvironment(r.Context(), serviceID, req.Environment, req.BaseURL, req.HealthEndpoint)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add environment", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, env, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) ListEnvironments(w http.ResponseWriter, r *http.Request, serviceID uuid.UUID) {
	envs, err := h.service.ListEnvironments(r.Context(), serviceID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list environments", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, envs, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) AddDependency(w http.ResponseWriter, r *http.Request, serviceID uuid.UUID) {
	var req addDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DependsOnServiceID == uuid.Nil {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "depends_on_service_id is required", httputil.RequestID(r.Context()))
		return
	}

	dep, err := h.service.AddDependency(r.Context(), serviceID, req.DependsOnServiceID, req.DependencyType, req.Critical)
	if err != nil {
		if errors.Is(err, domain.ErrSelfDependency) {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_DEPENDENCY", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add dependency", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, dep, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) ListDependencies(w http.ResponseWriter, r *http.Request, serviceID uuid.UUID) {
	deps, err := h.service.ListDependencies(r.Context(), serviceID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list dependencies", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, deps, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) AddHealthCheck(w http.ResponseWriter, r *http.Request, envID uuid.UUID) {
	var req addHealthCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "health check name is required", httputil.RequestID(r.Context()))
		return
	}

	hc, err := h.service.AddHealthCheck(r.Context(), envID, req.Name, req.Method, req.Path, req.ExpectedStatus, req.TimeoutMS, req.IntervalSeconds)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add health check", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, hc, httputil.RequestID(r.Context()))
}

func (h *RegistryHandler) ListHealthChecks(w http.ResponseWriter, r *http.Request, envID uuid.UUID) {
	hcs, err := h.service.ListHealthChecks(r.Context(), envID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list health checks", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, hcs, httputil.RequestID(r.Context()))
}
