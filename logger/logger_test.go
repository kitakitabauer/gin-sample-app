package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func resetLogger() {
	Sync()
	Log = nil
	atlReady = false
}

func TestInitSetsLogLevel(t *testing.T) {
	t.Cleanup(func() {
		resetLogger()
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

func TestSetLevel(t *testing.T) {
	defer resetLogger()

	if err := Init(Config{Env: "dev", Level: "info"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if err := SetLevel("error"); err != nil {
		t.Fatalf("set level failed: %v", err)
	}

	level, err := CurrentLevel()
	if err != nil {
		t.Fatalf("current level error: %v", err)
	}
	if level != zapcore.ErrorLevel {
		t.Fatalf("expected error level, got %s", level)
	}
}
