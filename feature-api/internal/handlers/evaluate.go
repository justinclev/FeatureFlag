package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/featureflags/feature-api/internal/models"
)

var (
	evalCtxPool = sync.Pool{
		New: func() any {
			return new(models.EvaluationContext)
		},
	}
	bufferPool = sync.Pool{
		New: func() any {
			return make([]byte, 32*1024)
		},
	}
)

func (h *Handler) evaluateFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if err := h.validateKey(key); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "request invalid")
		return
	}

	evalCtx := evalCtxPool.Get().(*models.EvaluationContext)
	// Zero out the entire struct safely to prevent cross-request leakage
	*evalCtx = models.EvaluationContext{}
	defer evalCtxPool.Put(evalCtx)

	if err := json.Unmarshal(body, evalCtx); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx, cancel := h.requestCtx(r)
	defer cancel()

	flag, err := h.repo.GetByKey(ctx, key)
	if err != nil {
		h.mapRepoError(w, err, "evaluate flag")
		return
	}

	result := h.evaluator.Evaluate(flag, *evalCtx)
	writeJSON(w, http.StatusOK, result)
}
