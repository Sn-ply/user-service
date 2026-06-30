package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/snaply/user-service/internal/service"
	"go.uber.org/zap"
)

type AuthHandler struct {
	auth service.AuthService
	log  *zap.Logger
}

func NewAuthHandler(auth service.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{auth: auth, log: log}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email, username, and password are required")
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			respondError(w, http.StatusConflict, "email already in use")
		case errors.Is(err, service.ErrUsernameTaken):
			respondError(w, http.StatusConflict, "username already taken")
		default:
			h.log.Error("register error", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, pair, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCreds) {
			respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.log.Error("login error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"user":          user,
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token required")
		return
	}

	pair, err := h.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrTokenInvalid) {
			respondError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		h.log.Error("refresh error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, pair)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token required")
		return
	}

	if err := h.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		h.log.Error("logout error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
