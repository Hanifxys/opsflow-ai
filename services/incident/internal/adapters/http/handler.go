package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/opsflow/common/httputil"
	"github.com/opsflow/common/middleware"
	"github.com/opsflow/incident-service/internal/application"
	"github.com/opsflow/incident-service/internal/domain"
	"github.com/opsflow/incident-service/internal/ports"
)

type IncidentHandler struct {
	service *application.IncidentService
}

func NewIncidentHandler(service *application.IncidentService) *IncidentHandler {
	return &IncidentHandler{service: service}
}

type createIncidentRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Severity    domain.Severity `json:"severity"`
	ServiceID   *uuid.UUID      `json:"service_id"`
}

type updateStatusRequest struct {
	Status  domain.Status `json:"status"`
	Message string        `json:"message"`
}

type resolveRequest struct {
	Notes string `json:"notes"`
}

type addCommentRequest struct {
	Content string `json:"content"`
}

type addEventRequest struct {
	EventType string                 `json:"event_type"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func (h *IncidentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", httputil.RequestID(r.Context()))
		return
	}

	if req.Title == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "title is required", httputil.RequestID(r.Context()))
		return
	}

	createdByStr := getActorID(r)
	createdBy, err := uuid.Parse(createdByStr)
	if err != nil {
		createdBy = uuid.Nil
	}

	inc, err := h.service.CreateIncident(r.Context(), req.Title, req.Description, req.Severity, req.ServiceID, createdBy)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create incident", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, inc, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := ports.ListFilter{}

	if s := q.Get("status"); s != "" {
		st := domain.Status(strings.ToUpper(s))
		if st.IsValid() {
			filter.Status = &st
		}
	}
	if s := q.Get("severity"); s != "" {
		sev := domain.Severity(strings.ToUpper(s))
		if sev.IsValid() {
			filter.Severity = &sev
		}
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	filter.Limit = limit
	filter.Offset = offset

	incidents, total, err := h.service.ListIncidents(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list incidents", httputil.RequestID(r.Context()))
		return
	}

	resp := map[string]interface{}{
		"incidents": incidents,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	}

	httputil.WriteSuccess(w, http.StatusOK, resp, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) Get(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	inc, err := h.service.GetIncident(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrIncidentNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Incident not found", httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get incident", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, inc, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", httputil.RequestID(r.Context()))
		return
	}

	actorIDStr := getActorID(r)
	actorID, _ := uuid.Parse(actorIDStr)

	inc, err := h.service.UpdateStatus(r.Context(), id, req.Status, actorID, req.Message)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		if errors.Is(err, domain.ErrIncidentNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Incident not found", httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update incident status", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, inc, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) Resolve(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req resolveRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	actorIDStr := getActorID(r)
	actorID, _ := uuid.Parse(actorIDStr)

	inc, err := h.service.ResolveIncident(r.Context(), id, actorID, req.Notes)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to resolve incident", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, inc, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) AddComment(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req addCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "content is required", httputil.RequestID(r.Context()))
		return
	}

	authorIDStr := getActorID(r)
	authorID, _ := uuid.Parse(authorIDStr)

	comment, err := h.service.AddComment(r.Context(), id, authorID, req.Content)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add comment", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, comment, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) ListComments(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	comments, err := h.service.ListComments(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list comments", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, comments, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) AddEvent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req addEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EventType == "" || req.Message == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "event_type and message are required", httputil.RequestID(r.Context()))
		return
	}

	actorIDStr := getActorID(r)
	actorID, _ := uuid.Parse(actorIDStr)

	event, err := h.service.AddEvent(r.Context(), id, req.EventType, req.Message, &actorID, req.Metadata)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add event", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, event, httputil.RequestID(r.Context()))
}

func (h *IncidentHandler) ListEvents(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	events, err := h.service.ListEvents(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list events", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, events, httputil.RequestID(r.Context()))
}

func getActorID(r *http.Request) string {
	if uid := middleware.UserID(r.Context()); uid != "" {
		return uid
	}
	return r.Header.Get("X-User-ID")
}
