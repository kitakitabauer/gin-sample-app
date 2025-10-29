package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"

	"github.com/kitakitabauer/gin-sample-app/config"
	"github.com/kitakitabauer/gin-sample-app/logger"
)

func setAdminAPIKey(t *testing.T, key string) func() {
	t.Helper()

	old := config.AppConfig
	config.AppConfig = &config.Config{APIKey: key}

	return func() {
		config.AppConfig = old
	}
}

func initLoggerForTest(t *testing.T, level string) {
	if err := logger.Init(logger.Config{Env: "dev", Level: level}); err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	t.Cleanup(logger.Sync)
}

func setupAdminRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewAdminHandler().RegisterRoutes(router)
	return router
}

func TestAdminHandler_GetLogLevel(t *testing.T) {
	t.Cleanup(setAdminAPIKey(t, "secret"))
	initLoggerForTest(t, "info")

	router := setupAdminRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/log-level", nil)
	req.Header.Set("X-API-Key", "secret")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body["level"] != "info" {
		t.Fatalf("expected level info, got %s", body["level"])
	}
}

func TestAdminHandler_UpdateLogLevel(t *testing.T) {
	t.Cleanup(setAdminAPIKey(t, "secret"))
	initLoggerForTest(t, "debug")

	router := setupAdminRouter(t)

	payload := []byte(`{"level":"error"}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/log-level", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	level, err := logger.CurrentLevel()
	if err != nil {
		t.Fatalf("failed to fetch current level: %v", err)
	}
	if level != zapcore.ErrorLevel {
		t.Fatalf("expected error level, got %s", level)
	}
}

func TestAdminHandler_UpdateLogLevel_InvalidPayload(t *testing.T) {
	t.Cleanup(setAdminAPIKey(t, "secret"))
	initLoggerForTest(t, "debug")

	router := setupAdminRouter(t)

	payload := []byte(`{"level":"invalid"}`)
	req := httptest.NewRequest(http.MethodPut, "/admin/log-level", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "secret")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("expected error message in response")
	}
}

func TestAdminHandler_Unauthorized(t *testing.T) {
	t.Cleanup(setAdminAPIKey(t, "secret"))
	initLoggerForTest(t, "info")

	router := setupAdminRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/log-level", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}
