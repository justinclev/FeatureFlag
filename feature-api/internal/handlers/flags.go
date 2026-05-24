package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
)

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.repo.Ready(ctx); err != nil {
		h.logger.Error("health check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listFlags(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var limit, offset int64 = 50, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 64); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	flags, err := h.repo.List(ctx, limit, offset)
	if err != nil {
		h.logger.ErrorContext(ctx, "list flags", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to fetch flags")
		return
	}
	writeJSON(w, http.StatusOK, flags)
}

func (h *Handler) createFlag(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // 1MB limit
	var req models.CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validateCreateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // 1MB limit
	var req models.UpdateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validateUpdateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

// validateCreateRequest ensures the flag and its rules are structurally sound.
func validateCreateRequest(req models.CreateFlagRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Key == "" {
		return errors.New("key is required")
	}
	for _, r := range req.Rules {
		if err := validateRule(r); err != nil {
			return err
		}
	}
	return nil
}

// validateUpdateRequest ensures the update fields are valid.
func validateUpdateRequest(req models.UpdateFlagRequest) error {
	if req.Rules != nil {
		for _, r := range *req.Rules {
			if err := validateRule(r); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRule(r models.Rule) error {
	if r.Type == "" {
		return errors.New("rule type is required")
	}
	switch r.Type {
	case models.RuleTypePercentage:
		if r.Config.Percentage == nil {
			return errors.New("percentage config missing 'percentage'")
		}
	case models.RuleTypeAttribute:
		if r.Config.AttributeKey == "" || r.Config.AttributeOp == "" {
			return errors.New("attribute config missing 'attributeKey' or 'attributeOp'")
		}
	case models.RuleTypeGradual:
		c := r.Config
		if c.StartAt == nil || c.EndAt == nil || c.StartPercent == nil || c.EndPercent == nil {
			return errors.New("gradual rollout config missing required fields")
		}
	}
	return nil
}
