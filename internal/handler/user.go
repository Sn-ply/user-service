package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/snaply/user-service/internal/model"
	"github.com/snaply/user-service/internal/service"
	"go.uber.org/zap"
)

type UserHandler struct {
	users service.UserService
	log   *zap.Logger
}

func NewUserHandler(users service.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{users: users, log: log}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, err := h.users.GetProfile(r.Context(), username)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("get profile error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respondJSON(w, http.StatusOK, user.ToPublicProfile())
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "missing or invalid user identity")
		return
	}

	var req struct {
		Bio       string `json:"bio"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.users.UpdateProfile(r.Context(), userID, service.UpdateProfileInput{
		Bio:       req.Bio,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("update profile error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Batch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.IDs) > 100 {
		respondError(w, http.StatusBadRequest, "too many ids (max 100)")
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid id: "+s)
			return
		}
		ids = append(ids, id)
	}

	users, err := h.users.GetByIDs(r.Context(), ids)
	if err != nil {
		h.log.Error("batch get users error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	profiles := make([]*model.PublicProfile, 0, len(users))
	for _, u := range users {
		profiles = append(profiles, u.ToPublicProfile())
	}
	respondJSON(w, http.StatusOK, profiles)
}

func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		respondError(w, http.StatusBadRequest, "q query parameter required")
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page, err := h.users.Search(r.Context(), q, cursor, limit)
	if err != nil {
		h.log.Error("search error", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":        page.Users,
		"next_cursor": page.NextCursor,
	})
}
