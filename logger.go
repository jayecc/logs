package logs

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"runtime"
	"strings"
)

type hookHandler func(r slog.Record)

type Hook struct {
	level  slog.Level
	handle hookHandler
}

func NewHook(level slog.Level, handle hookHandler) *Hook {
	return &Hook{
		level:  level,
		handle: handle,
	}
}

type hookTurn struct {
	r slog.Record
	h hookHandler
}

type Handler struct {
	slog.Handler
	hooks    []*Hook
	hookChan chan *hookTurn
}

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

func WrapHandler(handler slog.Handler, hooks ...*Hook) *Handler {
	h := &Handler{
		Handler:  handler,
		hooks:    hooks,
		hookChan: make(chan *hookTurn, 1<<10),
	}
	if len(hooks) > 0 {
		go h.hook()
	}
	return h
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h *Handler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.Handler = h.Handler.WithGroup(name)
	return &clone
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.Handler = h.Handler.WithAttrs(attrs)
	return &clone
}

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

	if r.Level == slog.LevelError {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		attrs = append(attrs, slog.Any(slog.SourceKey, slog.Source{Function: f.Function, File: f.File, Line: f.Line}))
	}

	if len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}

	for _, hook := range h.hooks {

		if len(h.hookChan) >= (1<<10)-1 {
			continue
		}

		if hook.level == r.Level {
			h.hookChan <- &hookTurn{h: hook.handle, r: r}
		}
	}

	return h.Handler.Handle(ctx, r)
}

func (h *Handler) hook() {
	for {
		select {
		case hook := <-h.hookChan:
			hook.h(hook.r)
		}
	}
}

//idx := strings.LastIndexByte(f.File, '/')
//if idx > 0 {
//	idx = strings.LastIndexByte(f.File[:idx], '/')
//	if idx > 0 {
//		f.File = f.File[idx+1:]
//	}
//}
//buf := make([]byte, 64<<10)
//n := runtime.Stack(buf, false)
//buf = buf[:n]
