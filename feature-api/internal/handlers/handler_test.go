package handlers_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/featureflags/feature-api/internal/handlers"
	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// mockRepo is an in-memory implementation of repository.FlagRepository for tests.
type mockRepo struct {
	flags []models.Flag
	err   error
}

func (m *mockRepo) List(_ context.Context) ([]models.Flag, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.flags == nil {
		return []models.Flag{}, nil
	}
	return m.flags, nil
}

func (m *mockRepo) GetByID(_ context.Context, id string) (*models.Flag, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i, f := range m.flags {
		if f.ID.Hex() == id {
			return &m.flags[i], nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockRepo) Create(_ context.Context, req models.CreateFlagRequest) (*models.Flag, error) {
	if m.err != nil {
		return nil, m.err
	}
	f := models.Flag{
		ID:          bson.NewObjectID(),
		Name:        req.Name,
		Enabled:     req.Enabled,
		Description: req.Description,
	}
	m.flags = append(m.flags, f)
	return &m.flags[len(m.flags)-1], nil
}

func (m *mockRepo) Update(_ context.Context, id string, req models.UpdateFlagRequest) (*models.Flag, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i, f := range m.flags {
		if f.ID.Hex() == id {
			if req.Name != nil {
				m.flags[i].Name = *req.Name
			}
			if req.Enabled != nil {
				m.flags[i].Enabled = *req.Enabled
			}
			if req.Description != nil {
				m.flags[i].Description = *req.Description
			}
			return &m.flags[i], nil
		}
	}
	return nil, repository.ErrNotFound
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	for i, f := range m.flags {
		if f.ID.Hex() == id {
			m.flags = append(m.flags[:i], m.flags[i+1:]...)
			return nil
		}
	}
	return repository.ErrNotFound
}

// newHandler creates a Handler with a quiet logger suitable for tests.
func newHandler(repo repository.FlagRepository) *handlers.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return handlers.New(repo, logger)
}

// serve registers routes and dispatches a single request, returning the recorder.
func serve(h *handlers.Handler, method, path, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHealth(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodGet, "/health", "")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestListFlags_ReturnsEmptySlice(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodGet, "/api/flags", "")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var flags []models.Flag
	if err := json.NewDecoder(rr.Body).Decode(&flags); err != nil {
		t.Fatal(err)
	}
	if len(flags) != 0 {
		t.Errorf("expected empty slice, got %d items", len(flags))
	}
}

func TestListFlags_ReturnsList(t *testing.T) {
	repo := &mockRepo{flags: []models.Flag{{ID: bson.NewObjectID(), Name: "my-flag", Enabled: true}}}
	h := newHandler(repo)
	rr := serve(h, http.MethodGet, "/api/flags", "")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var flags []models.Flag
	if err := json.NewDecoder(rr.Body).Decode(&flags); err != nil {
		t.Fatal(err)
	}
	if len(flags) != 1 {
		t.Errorf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Name != "my-flag" {
		t.Errorf("expected name=my-flag, got %q", flags[0].Name)
	}
}

func TestCreateFlag_MissingName(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPost, "/api/flags", `{"enabled":true}`)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateFlag_Success(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPost, "/api/flags", `{"name":"beta","enabled":false}`)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var flag models.Flag
	if err := json.NewDecoder(rr.Body).Decode(&flag); err != nil {
		t.Fatal(err)
	}
	if flag.Name != "beta" {
		t.Errorf("expected name=beta, got %q", flag.Name)
	}
}

func TestGetFlag_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodGet, "/api/flags/"+bson.NewObjectID().Hex(), "")

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteFlag_Success(t *testing.T) {
	id := bson.NewObjectID()
	repo := &mockRepo{flags: []models.Flag{{ID: id, Name: "to-delete"}}}
	h := newHandler(repo)
	rr := serve(h, http.MethodDelete, "/api/flags/"+id.Hex(), "")

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}
