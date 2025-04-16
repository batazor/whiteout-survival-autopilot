package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// TracedLogger расширяет slog и пишет в спан + лог с trace_id/span_id.
type TracedLogger struct {
	logger *slog.Logger
}

// NewTracedLogger создает новый TracedLogger с otelHandler.
func NewTracedLogger() *TracedLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	otelAware := &otelHandler{Handler: handler}

	return &TracedLogger{slog.New(otelAware)}
}

// Info отправляет сообщение уровня Info в лог и текущий span.
func (t *TracedLogger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(msg)
	t.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Warn отправляет сообщение уровня Warn в лог и текущий span.
func (t *TracedLogger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("WARN: " + msg)
	t.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// Error отправляет сообщение уровня Error в лог и текущий span.
func (t *TracedLogger) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(fmt.Errorf(msg))
	span.AddEvent("ERROR: " + msg)
	t.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// Debug отправляет сообщение уровня Debug в лог и текущий span.
func (t *TracedLogger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("DEBUG: " + msg)
	t.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// With создает новый TracedLogger с дополнительными атрибутами в виде пар (key, value).
func (t *TracedLogger) With(args ...any) *TracedLogger {
	// Преобразуем в []Attr, как делает slog.Logger.With()
	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue // можно panic/лог — но просто скипаем
		}
		attrs = append(attrs, slog.Any(key, args[i+1]))
	}

	handlerWithAttrs := t.logger.Handler().WithAttrs(attrs)
	return &TracedLogger{
		logger: slog.New(&otelHandler{Handler: handlerWithAttrs}),
	}
}
