package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/featureflags/feature-api/internal/models"
)

var evalCtxPool = sync.Pool{
	New: func() any {
		return &models.EvaluationContext{}
	},
}

func (h *Handler) evaluateFlag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.validateID(id); err != nil {
		h.mapRepoError(w, err, "evaluate flag")
		return
	}

	// Security: Limit request body size to 1MB to prevent OOM attacks.
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "request too large or invalid")
		return
	}

	evalCtx := evalCtxPool.Get().(*models.EvaluationContext)
	defer evalCtxPool.Put(evalCtx)

	// Performance: Manually reset fields to preserve map capacity.
	// clear() empties the map but keeps the underlying memory allocated.
	evalCtx.UserID = ""
	evalCtx.Country = ""
	evalCtx.State = ""
	evalCtx.City = ""
	evalCtx.ZipCode = ""
	if evalCtx.Attributes == nil {
		evalCtx.Attributes = make(map[string]any)
	} else {
		clear(evalCtx.Attributes)
	}

	if err := json.Unmarshal(body, evalCtx); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := h.requestCtx(r)
	defer cancel()

	flag, err := h.repo.GetByID(ctx, id)
	if err != nil {
		h.mapRepoError(w, err, "evaluate flag")
		return
	}

	result := h.evaluator.Evaluate(flag, *evalCtx)
	writeJSON(w, http.StatusOK, result)
}
