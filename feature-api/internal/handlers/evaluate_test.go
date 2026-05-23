package handlers_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/featureflags/feature-api/internal/handlers"
	"github.com/featureflags/feature-api/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEvaluateFlag_Success(t *testing.T) {
	id := bson.NewObjectID()
	repo := &mockRepo{flags: []models.Flag{{ID: id, Name: "my-flag", Enabled: true}}}
	eval := &mockEvaluator{result: models.EvaluationResult{Enabled: true, Reason: "matched rule: percentage"}}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	h := handlers.New(repo, logger, eval)

	rr := serve(h, http.MethodPost, "/api/flags/"+id.Hex()+"/evaluate", `{"userId":"user-1"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result models.EvaluationResult
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !result.Enabled {
		t.Errorf("expected enabled=true, got false")
	}
	if result.Reason != "matched rule: percentage" {
		t.Errorf("unexpected reason: %q", result.Reason)
	}
}

func TestEvaluateFlag_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPost, "/api/flags/"+bson.NewObjectID().Hex()+"/evaluate", `{"userId":"user-1"}`)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestEvaluateFlag_BadBody(t *testing.T) {
	id := bson.NewObjectID()
	repo := &mockRepo{flags: []models.Flag{{ID: id, Name: "my-flag", Enabled: true}}}
	h := newHandler(repo)
	rr := serve(h, http.MethodPost, "/api/flags/"+id.Hex()+"/evaluate", `not json`)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestEvaluateFlag_InvalidJSON(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPost, "/api/flags/id/evaluate", "invalid-json")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
