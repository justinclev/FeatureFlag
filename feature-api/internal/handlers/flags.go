package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
)

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.requestCtx(r)
	defer cancel()

	if err := h.repo.Ready(ctx); err != nil {
		h.logger.Error("health check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listFlags(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.requestCtx(r)
	defer cancel()

	limit := getQueryInt64(r, "limit", 50)
	offset := getQueryInt64(r, "offset", 0)

	// Security: Cap limit to prevent OOM
	if limit > 100 {
		limit = 100
	}
	// Security: Cap offset to prevent deep DB scan death
	if offset > 1000 {
		offset = 1000
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
	var req models.CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Business Logic: Apply defaults at handler level
	if req.RuleMatchStrategy == "" {
		req.RuleMatchStrategy = models.RuleMatchStrategyAny
	}

	if err := validateCreateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := h.requestCtx(r)
	defer cancel()

	flag, err := h.repo.Create(ctx, req)
	if err != nil {
		h.mapRepoError(w, err, "create flag")
		return
	}
	writeJSON(w, http.StatusCreated, flag)
}

func (h *Handler) getFlag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.validateID(id); err != nil {
		h.mapRepoError(w, err, "get flag")
		return
	}

	ctx, cancel := h.requestCtx(r)
	defer cancel()

	flag, err := h.repo.GetByID(ctx, id)
	if err != nil {
		h.mapRepoError(w, err, "get flag")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *Handler) updateFlag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.validateID(id); err != nil {
		h.mapRepoError(w, err, "update flag")
		return
	}

	var req models.UpdateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validateUpdateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := h.requestCtx(r)
	defer cancel()

	flag, err := h.repo.Update(ctx, id, req)
	if err != nil {
		h.mapRepoError(w, err, "update flag")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *Handler) deleteFlag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.validateID(id); err != nil {
		h.mapRepoError(w, err, "delete flag")
		return
	}

	ctx, cancel := h.requestCtx(r)
	defer cancel()

	if err := h.repo.Delete(ctx, id); err != nil {
		h.mapRepoError(w, err, "delete flag")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) mapRepoError(w http.ResponseWriter, err error, op string) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "flag not found")
	case errors.Is(err, repository.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "invalid id format")
	case errors.Is(err, repository.ErrNoFields):
		writeError(w, http.StatusBadRequest, "no fields to update")
	case errors.Is(err, repository.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "flag key already exists")
	default:
		h.logger.Error(op, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func getQueryInt64(r *http.Request, key string, fallback int64) int64 {
	val := r.URL.Query().Get(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func validateCreateRequest(req models.CreateFlagRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Key == "" {
		return errors.New("key is required")
	}
	if len(req.Key) > 64 {
		return errors.New("key length exceeds 64 chars")
	}
	for _, r := range req.Rules {
		if err := validateRule(r); err != nil {
			return err
		}
	}
	return nil
}

func validateUpdateRequest(req models.UpdateFlagRequest) error {
	if req.Key != nil && len(*req.Key) > 64 {
		return errors.New("key length exceeds 64 chars")
	}
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
		v, ok := r.Config["percentage"]
		if !ok || v == nil {
			return errors.New("percentage config missing 'percentage'")
		}
		p, ok := v.(float64)
		if !ok || p < 0 || p > 100 {
			return errors.New("percentage must be a number between 0 and 100")
		}
	case models.RuleTypeAttribute:
		ak, ok1 := r.Config["attributeKey"].(string)
		ao, ok2 := r.Config["attributeOp"].(string)
		if !ok1 || !ok2 || ak == "" || ao == "" {
			return errors.New("attribute config missing 'attributeKey' or 'attributeOp'")
		}
	case models.RuleTypeGradual:
		c := r.Config
		if c["startAt"] == nil || c["endAt"] == nil || c["startPercent"] == nil || c["endPercent"] == nil {
			return errors.New("gradual rollout config missing required fields")
		}
		sp, ok1 := c["startPercent"].(float64)
		ep, ok2 := c["endPercent"].(float64)
		if !ok1 || !ok2 || sp < 0 || sp > 100 || ep < 0 || ep > 100 {
			return errors.New("gradual percentages must be numbers between 0 and 100")
		}
	}
	return nil
}
