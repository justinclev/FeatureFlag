package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
)

// Handler holds application dependencies and exposes HTTP handler methods.
type Handler struct {
	repo      repository.FlagRepository
	logger    *slog.Logger
	evaluator FlagEvaluator
}

type FlagEvaluator interface {
	Evaluate(flag models.Flag, ctx models.EvaluationContext) models.EvaluationResult
}

// New constructs a Handler with the provided dependencies.
func New(repo repository.FlagRepository, logger *slog.Logger, evaluator FlagEvaluator) *Handler {
	return &Handler{repo: repo, logger: logger, evaluator: evaluator}
}

// RegisterRoutes registers all application routes on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /api/flags", h.listFlags)
	mux.HandleFunc("POST /api/flags", h.createFlag)
	mux.HandleFunc("GET /api/flags/{id}", h.getFlag)
	mux.HandleFunc("PATCH /api/flags/{id}", h.updateFlag)
	mux.HandleFunc("DELETE /api/flags/{id}", h.deleteFlag)
	mux.HandleFunc("POST /api/flags/{id}/evaluate", h.evaluateFlag)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
