package transport

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// captureHandler is a slog.Handler that records every record as flat text
// (message, attributes, groups) so tests can scan all log output.
type captureHandler struct {
	mu    sync.Mutex
	lines []string
}

func (c *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (c *captureHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Level.String() + " " + r.Message)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&b, " %s=%s", a.Key, a.Value.Resolve().String())
		return true
	})
	c.mu.Lock()
	c.lines = append(c.lines, b.String())
	c.mu.Unlock()
	return nil
}

func (c *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &attrHandler{parent: c, attrs: attrs}
}
func (c *captureHandler) WithGroup(string) slog.Handler { return c }

func (c *captureHandler) all() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return strings.Join(c.lines, "\n")
}

// attrHandler carries With attrs into the parent capture.
type attrHandler struct {
	parent *captureHandler
	attrs  []slog.Attr
}

func (a *attrHandler) Enabled(context.Context, slog.Level) bool { return true }
func (a *attrHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(a.attrs...)
	return a.parent.Handle(ctx, r)
}
func (a *attrHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &attrHandler{parent: a.parent, attrs: append(append([]slog.Attr{}, a.attrs...), attrs...)}
}
func (a *attrHandler) WithGroup(string) slog.Handler { return a }
