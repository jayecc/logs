package logs

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"strings"
)

type Handler struct {
	slog.Handler
}

// ParseLevel parses a level string.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WrapHandler wraps a slog.Handler.
func WrapHandler(handler slog.Handler) *Handler {
	h := &Handler{
		Handler: handler,
	}
	return h
}

// Enabled implements slog.Handler.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// WithGroup implements slog.Handler.
func (h *Handler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.Handler = h.Handler.WithGroup(name)
	return &clone
}

// WithAttrs implements slog.Handler.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.Handler = h.Handler.WithAttrs(attrs)
	return &clone
}

// Handle implements slog.Handler.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {

	if !h.Handler.Enabled(ctx, r.Level) {
		return h.Handler.Handle(ctx, r)
	}

	var attrs []slog.Attr

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		attrs = append(attrs, slog.String("trace", spanCtx.TraceID().String()))
	}

	if spanCtx.HasSpanID() {
		attrs = append(attrs, slog.String("span", spanCtx.SpanID().String()))
	}

	if len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}

	return h.Handler.Handle(ctx, r)
}
