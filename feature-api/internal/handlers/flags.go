package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
)

// Handler holds application dependencies and exposes HTTP handler methods.
type Handler struct {
	repo   repository.FlagRepository
	logger *slog.Logger
}

// New constructs a Handler with the provided dependencies.
func New(repo repository.FlagRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

// RegisterRoutes registers all application routes on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /api/flags", h.listFlags)
	mux.HandleFunc("POST /api/flags", h.createFlag)
	mux.HandleFunc("GET /api/flags/{id}", h.getFlag)
	mux.HandleFunc("PATCH /api/flags/{id}", h.updateFlag)
	mux.HandleFunc("DELETE /api/flags/{id}", h.deleteFlag)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listFlags(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	flags, err := h.repo.List(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "list flags", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch flags")
		return
	}
	writeJSON(w, http.StatusOK, flags)
}

func (h *Handler) createFlag(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	flag, err := h.repo.Create(ctx, req)
	if err != nil {
		h.logger.ErrorContext(ctx, "create flag", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create flag")
		return
	}
	writeJSON(w, http.StatusCreated, flag)
}

func (h *Handler) getFlag(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	flag, err := h.repo.GetByID(ctx, r.PathValue("id"))
	if err != nil {
		h.mapRepoError(w, err, "get flag")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *Handler) updateFlag(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	flag, err := h.repo.Update(ctx, r.PathValue("id"), req)
	if err != nil {
		h.mapRepoError(w, err, "update flag")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *Handler) deleteFlag(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := h.repo.Delete(ctx, r.PathValue("id")); err != nil {
		h.mapRepoError(w, err, "delete flag")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// mapRepoError translates repository sentinel errors to HTTP responses.
func (h *Handler) mapRepoError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "flag not found")
	case errors.Is(err, repository.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "invalid id")
	case errors.Is(err, repository.ErrNoFields):
		writeError(w, http.StatusBadRequest, "no fields to update")
	default:
		h.logger.Error(op, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
