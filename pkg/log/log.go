package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

const (
	LoggerKey = "Logger"
)

// 1. 核心 Logger 接口（你项目所有插件只依赖这个）
type Logger interface {
	// 基础日志
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	// 结构化日志（带字段，插件必备）
	WithField(key string, value any) Logger
	WithFields(fields map[string]any) Logger

	// 上下文日志（trace_id / user_id）
	WithContext(ctx context.Context) Logger
}

// 2. 实现类（基于官方 slog）
type Base struct {
	logger *slog.Logger
}

// 确保实现 Logger 接口
var _ Logger = (*Base)(nil)

// 3. 构造函数（你原来的 New()）
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

// ------------------------------
// 日志方法实现
// ------------------------------
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

// ------------------------------
// 结构化字段
// ------------------------------
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

// ------------------------------
// 上下文支持
// ------------------------------
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

// ------------------------------
// 工具：简化日志里的文件路径
// ------------------------------
func simplifySource(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*runtime.Frame)
		a.Value = slog.StringValue(filepath.Base(source.File))
	}
	return a
}
