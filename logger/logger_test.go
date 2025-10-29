package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestInitSetsLogLevel(t *testing.T) {
	t.Cleanup(func() {
		Sync()
		Log = nil
	})

	err := Init(Config{Env: "prd", Level: "warn", Service: "api"})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if Log == nil {
		t.Fatalf("expected logger to be initialised")
	}

	if Log.Core().Enabled(zapcore.InfoLevel) {
		t.Fatalf("info level should be disabled when configured for warn")
	}
	if !Log.Core().Enabled(zapcore.WarnLevel) {
		t.Fatalf("warn level should be enabled")
	}
}

func TestInitInvalidLevel(t *testing.T) {
	if err := Init(Config{Env: "dev", Level: "verbose"}); err == nil {
		t.Fatalf("expected error for invalid log level")
	}
}

func TestParseLevelDefault(t *testing.T) {
	lvl, err := parseLevel("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lvl != zapcore.DebugLevel {
		t.Fatalf("expected default debug level, got %s", lvl)
	}
}
