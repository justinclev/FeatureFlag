package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
)

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
