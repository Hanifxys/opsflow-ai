package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/opsflow/auth-service/internal/application"
	"github.com/opsflow/auth-service/internal/domain"
	"github.com/opsflow/common/httputil"
	"github.com/opsflow/common/middleware"
)

type AuthHandler struct {
	service *application.AuthService
}

func NewAuthHandler(service *application.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type registerRequest struct {
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	DisplayName string   `json:"display_name"`
	Roles       []string `json:"roles"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", httputil.RequestID(r.Context()))
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", httputil.RequestID(r.Context()))
		return
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Email, password, and display_name are required", httputil.RequestID(r.Context()))
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password, req.DisplayName, req.Roles)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			httputil.WriteError(w, http.StatusConflict, "EMAIL_EXISTS", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user", httputil.RequestID(r.Context()))
		return
	}

	httputil.WriteSuccess(w, http.StatusCreated, user, httputil.RequestID(r.Context()))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", httputil.RequestID(r.Context()))
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", httputil.RequestID(r.Context()))
		return
	}

	accessToken, refreshToken, user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httputil.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		if errors.Is(err, domain.ErrUserInactive) {
			httputil.WriteError(w, http.StatusForbidden, "USER_INACTIVE", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to log in", httputil.RequestID(r.Context()))
		return
	}

	resp := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	}

	httputil.WriteSuccess(w, http.StatusOK, resp, httputil.RequestID(r.Context()))
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", httputil.RequestID(r.Context()))
		return
	}

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", httputil.RequestID(r.Context()))
		return
	}

	if req.RefreshToken == "" {
		httputil.WriteError(w, http.StatusBadRequest, "VALIDATION_FAILED", "refresh_token is required", httputil.RequestID(r.Context()))
		return
	}

	accessToken, refreshToken, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRefreshToken) {
			httputil.WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", err.Error(), httputil.RequestID(r.Context()))
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to refresh token", httputil.RequestID(r.Context()))
		return
	}

	resp := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	httputil.WriteSuccess(w, http.StatusOK, resp, httputil.RequestID(r.Context()))
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", httputil.RequestID(r.Context()))
		return
	}

	userIDStr := middleware.UserID(r.Context())
	if userIDStr == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User context missing", httputil.RequestID(r.Context()))
		return
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID format", httputil.RequestID(r.Context()))
		return
	}

	user, roles, perms, err := h.service.GetUserProfile(r.Context(), uid)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "USER_NOT_FOUND", "User profile not found", httputil.RequestID(r.Context()))
		return
	}

	resp := map[string]interface{}{
		"user":        user,
		"roles":       roles,
		"permissions": perms,
	}

	httputil.WriteSuccess(w, http.StatusOK, resp, httputil.RequestID(r.Context()))
}
