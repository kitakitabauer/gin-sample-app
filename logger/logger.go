package logger

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines how the global logger should be initialised.
type Config struct {
	Env     string
	Level   string
	Service string
}

var (
	// Log is the singleton zap.Logger used across the application.
	Log *zap.Logger

	mu sync.Mutex

	atl      zap.AtomicLevel
	atlReady bool
)

// Init builds a zap logger using the provided configuration.
// Callers should ensure Init is invoked once on application start.
func Init(cfg Config) error {
	mu.Lock()
	defer mu.Unlock()

	level, err := parseLevel(cfg.Level)
	if err != nil {
		return err
	}

	env := strings.ToLower(strings.TrimSpace(cfg.Env))
	if env == "" {
		env = "dev"
	}

	atomicLevel := zap.NewAtomicLevelAt(level)
	zapCfg := baseZapConfig(env)
	zapCfg.Level = atomicLevel
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.MessageKey = "message"
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.StacktraceKey = "stacktrace"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if zapCfg.InitialFields == nil {
		zapCfg.InitialFields = make(map[string]interface{})
	}
	zapCfg.InitialFields["env"] = env
	if service := strings.TrimSpace(cfg.Service); service != "" {
		zapCfg.InitialFields["service"] = service
	}

	logger, err := zapCfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	if Log != nil {
		_ = Log.Sync()
	}
	Log = logger
	atl = atomicLevel
	atlReady = true
	return nil
}

// Sync flushes any buffered log entries. It is safe to call multiple times.
func Sync() {
	mu.Lock()
	defer mu.Unlock()

	if Log == nil {
		return
	}
	_ = Log.Sync()
}

// SetLevel updates the runtime log level using Zap's atomic level.
func SetLevel(level string) error {
	mu.Lock()
	defer mu.Unlock()

	if !atlReady {
		return errors.New("logger not initialised")
	}

	parsed, err := parseLevel(level)
	if err != nil {
		return err
	}

	atl.SetLevel(parsed)
	return nil
}

// CurrentLevel returns the active log level.
func CurrentLevel() (zapcore.Level, error) {
	mu.Lock()
	defer mu.Unlock()

	if !atlReady {
		return zapcore.InfoLevel, errors.New("logger not initialised")
	}
	return atl.Level(), nil
}

func baseZapConfig(env string) zap.Config {
	switch env {
	case "prd", "production", "prod":
		return zap.NewProductionConfig()
	default:
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.DisableStacktrace = true
		return cfg
	}
}

func parseLevel(level string) (zapcore.Level, error) {
	trimmed := strings.ToLower(strings.TrimSpace(level))
	if trimmed == "" {
		return zapcore.DebugLevel, nil
	}

	lvl, err := zapcore.ParseLevel(trimmed)
	if err != nil {
		return zapcore.DebugLevel, fmt.Errorf("invalid log level %q: %w", level, err)
	}
	return lvl, nil
}
