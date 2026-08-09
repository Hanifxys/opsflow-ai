package http

import (
	"net/http"

	"github.com/opsflow/common/middleware"
)

func RegisterRoutes(mux *http.ServeMux, handler *AIHandler, jwtSecret string) {
	authMw := middleware.AuthMiddleware(jwtSecret)

	// Conversations
	mux.Handle("POST /api/v1/ai/conversations", authMw(middleware.RequirePermission("ai:use")(http.HandlerFunc(handler.CreateConversation))))
	mux.Handle("GET /api/v1/ai/conversations", authMw(middleware.RequirePermission("ai:use")(http.HandlerFunc(handler.ListConversations))))
	mux.Handle("POST /api/v1/ai/conversations/{id}/messages", authMw(middleware.RequirePermission("ai:use")(http.HandlerFunc(handler.SendMessage))))

	// Human Approvals Workflow
	mux.Handle("GET /api/v1/ai/approvals", authMw(middleware.RequirePermission("ai:use")(http.HandlerFunc(handler.ListPendingApprovals))))
	mux.Handle("POST /api/v1/ai/approvals/{id}/approve", authMw(middleware.RequirePermission("ai:execute")(http.HandlerFunc(handler.ApproveAction))))
	mux.Handle("POST /api/v1/ai/approvals/{id}/reject", authMw(middleware.RequirePermission("ai:execute")(http.HandlerFunc(handler.RejectAction))))

	// RAG Knowledge Ingestion
	mux.Handle("POST /api/v1/ai/knowledge", authMw(middleware.RequirePermission("ai:use")(http.HandlerFunc(handler.IngestKnowledge))))
}
