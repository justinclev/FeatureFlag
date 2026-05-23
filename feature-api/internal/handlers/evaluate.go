package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/featureflags/feature-api/internal/models"
)

func (h *Handler) evaluateFlag(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var evalCtx models.EvaluationContext
	if err := json.Unmarshal(body, &evalCtx); err != nil {
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

	result := h.evaluator.Evaluate(flag, evalCtx)

	// Performance: Manually write JSON to bypass reflection overhead in json.Encoder
	// for the high-throughput evaluation hot path (10k+ RPS goal).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"enabled":%t,"reason":"%s"}`, result.Enabled, result.Reason)
}
