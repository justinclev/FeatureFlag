package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

func (h *Handler) evaluateFlag(w http.ResponseWriter, r *http.Request) {
	var evalCtx models.EvaluationContext
	if err := json.NewDecoder(r.Body).Decode(&evalCtx); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	flag, err := h.repo.GetByID(ctx, r.PathValue("id"))
	if err != nil {
		h.mapRepoError(w, err, "evaluate flag")
		return
	}

	result := h.evaluator.Evaluate(*flag, evalCtx)
	writeJSON(w, http.StatusOK, result)
}
