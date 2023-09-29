package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// loggerZap wrapper of Zap logger.
type loggerZap struct {
	zap *zap.Logger
}

// CreateZapLogger create method for zap logger.
func CreateZapLogger(level string) (BaseLogger, error) {
	cfg, err := initConfig(level)
	if err != nil {
		return nil, err
	}

	log, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("create zap loggerZap^ %w", err)
	}

	return &loggerZap{zap: log}, nil
}

// Info generates 'info' level log.
func (l *loggerZap) Info(msg string, fields ...interface{}) {
	l.zap.Info(fmt.Sprintf(msg, fields...))
}

// Printf interface for kafka's implementation.
func (l *loggerZap) Printf(msg string, fields ...interface{}) {
	l.Info(msg, fields...)
}

// initConfig method that initializes logger.
func initConfig(level string) (zap.Config, error) {
	cfg := zap.NewProductionConfig()

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		emptyConfig := zap.Config{}
		return emptyConfig, fmt.Errorf("init zap loggerZap config: %w", err)
	}
	cfg.Level = lvl
	cfg.DisableCaller = true

	return cfg, nil
}

// Sync flushing any buffered log entries.
func (l *loggerZap) Sync() {
	err := l.zap.Sync()
	if err != nil {
		panic(err)
	}
}

// LogHandler handler for requests logging.
func (l *loggerZap) LogHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		loggingWriter := createResponseWriter(w)
		h.ServeHTTP(loggingWriter, r)
		duration := time.Since(start)

		l.zap.Info("new incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", loggingWriter.getResponseData().status),
			zap.String("content-type", loggingWriter.Header().Get("Content-Type")),
			zap.String("content-encoding", loggingWriter.Header().Get("Content-Encoding")),
			zap.String("HashSHA256", loggingWriter.Header().Get("HashSHA256")),
			zap.Duration("duration", duration),
			zap.Int("size", loggingWriter.getResponseData().size),
		)
	})
}
