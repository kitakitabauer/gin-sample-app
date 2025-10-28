package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kitakitabauer/gin-sample-app/handler"
	"github.com/kitakitabauer/gin-sample-app/model"
	"github.com/kitakitabauer/gin-sample-app/repository"
	"github.com/kitakitabauer/gin-sample-app/service"
)

func setupTestRouter() *gin.Engine {
	repo := repository.NewInMemoryPostRepository()
	postService := service.NewPostService(repo)
	postHandler := handler.NewPostHandler(postService)

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	postHandler.RegisterRoutes(router)

	return router
}

func TestHealthz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestPostEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupTestRouter()

	createPayload := map[string]string{
		"title":   "How to use Gin",
		"content": "Sample content for the post.",
		"author":  "Alice",
	}
	bodyBytes, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var created model.Post
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode created post: %v", err)
	}

	if created.ID == 0 || created.Title != createPayload["title"] || created.Content != createPayload["content"] || created.Author != createPayload["author"] {
		t.Fatalf("unexpected created post: %+v", created)
	}

	updatePayload := map[string]string{
		"title":  "Updated Gin Guide",
		"author": "Bob",
	}
	bodyBytes, err = json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("failed to marshal update payload: %v", err)
	}

	req = httptest.NewRequest(http.MethodPatch, "/posts/"+strconv.FormatInt(created.ID, 10), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var updated model.Post
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("failed to decode updated post: %v", err)
	}

	if updated.Title != updatePayload["title"] || updated.Author != updatePayload["author"] || updated.Content != created.Content {
		t.Fatalf("unexpected updated post: %+v", updated)
	}

	req = httptest.NewRequest(http.MethodGet, "/posts", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var list struct {
		Posts []model.Post `json:"posts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}

	if len(list.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(list.Posts))
	}

	req = httptest.NewRequest(http.MethodGet, "/posts/"+strconv.FormatInt(created.ID, 10), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var fetched model.Post
	if err := json.Unmarshal(rec.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("failed to decode fetched post: %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf("expected fetched ID %d, got %d", created.ID, fetched.ID)
	}

	req = httptest.NewRequest(http.MethodDelete, "/posts/"+strconv.FormatInt(created.ID, 10), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/posts/"+strconv.FormatInt(created.ID, 10), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/posts", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode list after delete: %v", err)
	}

	if len(list.Posts) != 0 {
		t.Fatalf("expected no posts after delete, got %d", len(list.Posts))
	}
}
