package logg

import (
	"bytes"
	"context"
	"log/slog"
	"runtime"
	"sync"

	PM "thornol/internal/protos"
)

type GolainHandler struct {
	h slog.Handler
	b *bytes.Buffer
	m *sync.Mutex
}

func NewGolainLogger(handler slog.Handler, levels []slog.Level) *GolainHandler {
	return &GolainHandler{
		h: handler,
	}
}

func (h *GolainHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *GolainHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &GolainHandler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m}
}

func (h *GolainHandler) WithGroup(name string) slog.Handler {
	return &GolainHandler{h: h.h.WithGroup(name), b: h.b, m: h.m}
}

func (h *GolainHandler) Handle(ctx context.Context, r slog.Record) error {
	l := PM.PLog{}
	switch r.Level {
	case slog.LevelDebug, slog.LevelInfo:
	case slog.LevelWarn, slog.LevelError:
		l.Level = PM.Level(r.Level)
		l.Message = r.Message
		l.TimeMs = uint32(r.Time.UnixMilli())
		l.Function = runtime.FuncForPC(r.PC).Name()
	}

	if err := h.h.Handle(ctx, r); err != nil {
		return err
	}

	// attempt to post to the server

	return nil
}
