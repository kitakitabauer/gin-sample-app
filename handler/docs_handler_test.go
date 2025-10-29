package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocsHandler_ServeYAML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewDocsHandler().RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if ct := resp.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Fatalf("expected content-type application/yaml, got %s", ct)
	}
	if len(resp.Body.Bytes()) == 0 {
		t.Fatal("expected body to be non-empty")
	}
}

func TestDocsHandler_ServeSwaggerUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewDocsHandler().RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/docs/swagger", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if ct := resp.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("expected content-type text/html; charset=utf-8, got %s", ct)
	}
	if !bytes.Contains(resp.Body.Bytes(), []byte("SwaggerUIBundle")) {
		t.Fatalf("expected body to contain SwaggerUIBundle")
	}
}

func TestDocsHandler_ServeRedoc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewDocsHandler().RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/docs/redoc", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if ct := resp.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("expected content-type text/html; charset=utf-8, got %s", ct)
	}
	if !bytes.Contains(resp.Body.Bytes(), []byte("Redoc.init")) {
		t.Fatalf("expected body to contain Redoc.init")
	}
}
