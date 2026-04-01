package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger

	WithContext(ctx context.Context) Logger
}

type Base struct {
	logger *slog.Logger
}

var _ Logger = (*Base)(nil)

func New(devMode bool) Logger {
	var handler slog.Handler

	if devMode {
		// 开发：文本格式，带文件行号，易读
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: simplifySource,
		})
	} else {
		// 生产：JSON 格式，ELK 友好
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})
	}

	return &Base{logger: slog.New(handler)}
}

func (b *Base) Debug(msg string, args ...any) {
	b.logger.Debug(msg, args...)
}

func (b *Base) Info(msg string, args ...any) {
	b.logger.Info(msg, args...)
}

func (b *Base) Warn(msg string, args ...any) {
	b.logger.Warn(msg, args...)
}

func (b *Base) Error(msg string, args ...any) {
	b.logger.Error(msg, args...)
}

func (b *Base) WithField(key string, value any) Logger {
	return &Base{
		logger: b.logger.With(slog.Any(key, value)),
	}
}

func (b *Base) WithFields(fields map[string]any) Logger {
	attrs := make([]any, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &Base{
		logger: b.logger.With(attrs...),
	}
}

type ctxKey struct{}

func (b *Base) WithContext(ctx context.Context) Logger {
	if traceID, ok := ctx.Value(ctxKey{}).(string); ok {
		return b.WithField("trace_id", traceID)
	}
	return b
}

// WithTraceID 给 ctx 放入 trace
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, traceID)
}

func simplifySource(_ []string, a slog.Attr) slog.Attr {
	// 只处理 source 字段
	if a.Key != slog.SourceKey {
		return a
	}

	// 安全类型断言（避免 panic）
	source, ok := a.Value.Any().(*runtime.Frame)
	if !ok {
		return a
	}

	// 返回新的 Attr：只保留文件名 + 行号
	file := filepath.Base(source.File)
	return slog.String(slog.SourceKey, file)
}

var (
	loggerInstance = New(false)
)

func Default() Logger {
	return loggerInstance
}
