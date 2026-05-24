package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/featureflags/feature-api/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEvaluateFlag_Match(t *testing.T) {
	id := bson.NewObjectID()
	key := "test-flag"
	repo := &mockRepo{flags: []models.Flag{{ID: id, Key: key, Enabled: true}}}
	h := newHandler(repo)

	// Mock evaluator result
	h.evaluator.(*mockEvaluator).result = models.EvaluationResult{Enabled: true, Reason: "match"}

	body := `{"userId":"user-1"}`
	rr := serve(h, http.MethodPost, "/api/flags/"+key+"/evaluate", body)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var res models.EvaluationResult
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if !res.Enabled {
		t.Error("expected enabled=true")
	}
}

func TestEvaluateFlag_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{})
	body := `{"userId":"user-1"}`
	rr := serve(h, http.MethodPost, "/api/flags/missing-key/evaluate", body)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestEvaluateFlag_InvalidJSON(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPost, "/api/flags/some-key/evaluate", "invalid-json")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestEvaluateFlag_KeyTooLong(t *testing.T) {
	h := newHandler(&mockRepo{})
	key := strings.Repeat("a", 65)
	rr := serve(h, http.MethodPost, "/api/flags/"+key+"/evaluate", `{"userId":"u1"}`)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for long key, got %d", rr.Code)
	}
}

func TestEvaluateFlag_RepoError(t *testing.T) {
	repo := &mockRepo{err: errors.New("fail")}
	h := newHandler(repo)
	rr := serve(h, http.MethodPost, "/api/flags/key/evaluate", `{"userId":"u1"}`)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
