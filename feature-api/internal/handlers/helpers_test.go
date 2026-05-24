package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// mockRepo is an in-memory implementation of repository.FlagRepository for tests.
type mockRepo struct {
	flags []models.Flag
	err   error
}

func (m *mockRepo) List(_ context.Context, limit, offset int64) ([]models.Flag, error) {
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

func (m *mockRepo) GetByKey(_ context.Context, key string) (*models.Flag, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i, f := range m.flags {
		if f.Key == key {
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
		Key:         req.Key,
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

func (m *mockRepo) Ready(_ context.Context) error {
	return m.err
}

type mockEvaluator struct {
	result models.EvaluationResult
}

func (m *mockEvaluator) Evaluate(_ *models.Flag, _ models.EvaluationContext) models.EvaluationResult {
	return m.result
}

// newHandler creates a Handler with a quiet logger suitable for tests.
func newHandler(repo repository.FlagRepository) *Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return New(repo, logger, &mockEvaluator{}, 5*time.Second)
}

// serve registers routes and dispatches a single request, returning the recorder.
func serve(h *Handler, method, path, body string) *httptest.ResponseRecorder {
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
