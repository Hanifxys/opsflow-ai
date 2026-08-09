package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/opsflow/ai-gateway/internal/application"
	"github.com/opsflow/common/httputil"
)

type AIHandler struct {
	service *application.AIService
}

func NewAIHandler(service *application.AIService) *AIHandler {
	return &AIHandler{service: service}
}

type CreateConversationRequest struct {
	Title string `json:"title"`
}

func (h *AIHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User identity missing", httputil.RequestID(r.Context()))
		return
	}

	var req CreateConversationRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	conv, err := h.service.CreateConversation(r.Context(), userID, req.Title)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, conv, httputil.RequestID(r.Context()))
}

func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User identity missing", httputil.RequestID(r.Context()))
		return
	}

	list, err := h.service.ListConversations(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, map[string]interface{}{"conversations": list}, httputil.RequestID(r.Context()))
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

func (h *AIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	convIDStr := r.PathValue("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Invalid conversation ID", httputil.RequestID(r.Context()))
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Message content is required", httputil.RequestID(r.Context()))
		return
	}

	msg, approvals, err := h.service.SendMessage(r.Context(), convID, userID, req.Content)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "AI_ERROR", err.Error(), httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"message":   msg,
		"approvals": approvals,
	}, httputil.RequestID(r.Context()))
}

func (h *AIHandler) ListPendingApprovals(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListPendingApprovals(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), httputil.RequestID(r.Context()))
		return
	}
	httputil.WriteSuccess(w, http.StatusOK, map[string]interface{}{"approvals": list}, httputil.RequestID(r.Context()))
}

func (h *AIHandler) ApproveAction(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	approverID, _ := uuid.Parse(userIDStr)

	approvalIDStr := r.PathValue("id")
	approvalID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Invalid approval ID", httputil.RequestID(r.Context()))
		return
	}

	output, err := h.service.ApproveToolAction(r.Context(), approvalID, approverID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "APPROVAL_FAILED", err.Error(), httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status": "APPROVED",
		"result": output,
	}, httputil.RequestID(r.Context()))
}

func (h *AIHandler) RejectAction(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	approverID, _ := uuid.Parse(userIDStr)

	approvalIDStr := r.PathValue("id")
	approvalID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Invalid approval ID", httputil.RequestID(r.Context()))
		return
	}

	if err := h.service.RejectToolAction(r.Context(), approvalID, approverID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "REJECTION_FAILED", err.Error(), httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusOK, map[string]interface{}{"status": "REJECTED"}, httputil.RequestID(r.Context()))
}

type IngestKnowledgeRequest struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *AIHandler) IngestKnowledge(w http.ResponseWriter, r *http.Request) {
	var req IngestKnowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.Content == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Title and content required", httputil.RequestID(r.Context()))
		return
	}

	doc, err := h.service.IngestKnowledge(r.Context(), req.Title, req.Content, req.Metadata)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, doc, httputil.RequestID(r.Context()))
}
