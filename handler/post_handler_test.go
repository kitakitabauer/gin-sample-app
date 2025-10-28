package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kitakitabauer/gin-sample-app/config"
	"github.com/kitakitabauer/gin-sample-app/model"
	"github.com/kitakitabauer/gin-sample-app/repository"
	"github.com/kitakitabauer/gin-sample-app/service"
)

func setAPIKeyForTest(t *testing.T, key string) func() {
	t.Helper()

	original := config.AppConfig
	config.AppConfig = &config.Config{
		APIKey: key,
	}

	return func() {
		config.AppConfig = original
	}
}

func setupTestRouter(t *testing.T) (*gin.Engine, *repository.InMemoryPostRepository) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	repo := repository.NewInMemoryPostRepository()
	svc := service.NewPostService(repo)
	handler := NewPostHandler(svc)

	router := gin.New()
	handler.RegisterRoutes(router)

	return router, repo
}

func TestPostHandler_CreatePost_Success(t *testing.T) {
	t.Cleanup(setAPIKeyForTest(t, ""))

	router, _ := setupTestRouter(t)

	body := map[string]string{
		"title":   "Test Title",
		"content": "Test Content",
		"author":  "Alice",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var created model.Post
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}

	if created.Title != body["title"] || created.Content != body["content"] || created.Author != body["author"] {
		t.Fatalf("unexpected created post: %+v", created)
	}
}

func TestPostHandler_CreatePost_WithAPIKey(t *testing.T) {
	t.Cleanup(setAPIKeyForTest(t, "secret"))

	router, _ := setupTestRouter(t)

	body := map[string]string{
		"title":   "Protected",
		"content": "Content",
		"author":  "Alice",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestPostHandler_CreatePost_Unauthorized(t *testing.T) {
	t.Cleanup(setAPIKeyForTest(t, "secret"))

	router, _ := setupTestRouter(t)

	payload := []byte(`{"title":"Protected","content":"Content","author":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestPostHandler_CreatePost_ValidationError(t *testing.T) {
	t.Cleanup(setAPIKeyForTest(t, ""))

	router, _ := setupTestRouter(t)

	payload := []byte(`{"title":"","content":"c","author":"a"}`)
	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}

	if body["error"] != service.ErrTitleRequired.Error() {
		t.Fatalf("expected error %q, got %q", service.ErrTitleRequired.Error(), body["error"])
	}
}

func TestPostHandler_UpdatePost_NoFields(t *testing.T) {
	t.Cleanup(setAPIKeyForTest(t, ""))

	router, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodPatch, "/posts/1", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}

	if body["error"] != service.ErrNoFieldsToUpdate.Error() {
		t.Fatalf("expected error %q, got %q", service.ErrNoFieldsToUpdate.Error(), body["error"])
	}
}

func TestPostHandler_UpdatePost_Success(t *testing.T) {
	t.Cleanup(setAPIKeyForTest(t, "secret"))

	router, _ := setupTestRouter(t)

	// create initial post
	createPayload := []byte(`{"title":"original","content":"content","author":"Alice"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-API-Key", "secret")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("failed to create post, status %d body %s", createRec.Code, createRec.Body.String())
	}

	updatePayload := []byte(`{"title":"updated title"}`)
	req := httptest.NewRequest(http.MethodPatch, "/posts/1", bytes.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var updated model.Post
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}

	if updated.Title != "updated title" {
		t.Fatalf("expected title to be updated, got %q", updated.Title)
	}
}

func TestPostHandler_DeletePost_NotFound(t *testing.T) {
	t.Cleanup(setAPIKeyForTest(t, "secret"))

	router, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/posts/99", nil)
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}

	if body["error"] != "post not found" {
		t.Fatalf("expected error 'post not found', got %q", body["error"])
	}
}
