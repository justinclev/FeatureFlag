package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/featureflags/feature-api/internal/models"
	"github.com/featureflags/feature-api/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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

func TestListFlags_RepoError(t *testing.T) {
	repo := &mockRepo{err: errors.New("db error")}
	h := newHandler(repo)
	rr := serve(h, http.MethodGet, "/api/flags", "")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestGetFlag_Success(t *testing.T) {
	id := bson.NewObjectID()
	repo := &mockRepo{flags: []models.Flag{{ID: id, Name: "get-me"}}}
	h := newHandler(repo)
	rr := serve(h, http.MethodGet, "/api/flags/"+id.Hex(), "")

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var flag models.Flag
	if err := json.NewDecoder(rr.Body).Decode(&flag); err != nil {
		t.Fatal(err)
	}
	if flag.Name != "get-me" {
		t.Errorf("expected name=get-me, got %q", flag.Name)
	}
}

func TestUpdateFlag_Success(t *testing.T) {
	id := bson.NewObjectID()
	repo := &mockRepo{flags: []models.Flag{{ID: id, Name: "old-name"}}}
	h := newHandler(repo)
	rr := serve(h, http.MethodPatch, "/api/flags/"+id.Hex(), `{"name":"new-name"}`)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var flag models.Flag
	if err := json.NewDecoder(rr.Body).Decode(&flag); err != nil {
		t.Fatal(err)
	}
	if flag.Name != "new-name" {
		t.Errorf("expected name=new-name, got %q", flag.Name)
	}
}

func TestUpdateFlag_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPatch, "/api/flags/"+bson.NewObjectID().Hex(), `{"name":"new-name"}`)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteFlag_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodDelete, "/api/flags/"+bson.NewObjectID().Hex(), "")

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestCreateFlag_InvalidJSON(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPost, "/api/flags", `invalid-json`)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestUpdateFlag_InvalidJSON(t *testing.T) {
	h := newHandler(&mockRepo{})
	rr := serve(h, http.MethodPatch, "/api/flags/"+bson.NewObjectID().Hex(), `invalid-json`)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestMapRepoError_InvalidID(t *testing.T) {
	repo := &mockRepo{err: repository.ErrInvalidID}
	h := newHandler(repo)
	rr := serve(h, http.MethodGet, "/api/flags/"+bson.NewObjectID().Hex(), "")
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestMapRepoError_NoFields(t *testing.T) {
	repo := &mockRepo{err: repository.ErrNoFields}
	h := newHandler(repo)
	rr := serve(h, http.MethodPatch, "/api/flags/"+bson.NewObjectID().Hex(), `{"name":"new"}`)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
