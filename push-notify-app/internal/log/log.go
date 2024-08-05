package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelNone
)

type Logger struct {
	logLevel LogLevel
	out      io.Writer
}

func NewLogger(logLevel string, out io.Writer) *Logger {
	var level LogLevel
	switch logLevel {
	case "trace":
		level = LevelTrace
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	case "warn":
		level = LevelWarn
	case "error":
		level = LevelError
	case "none":
		level = LevelNone
	default:
		level = LevelDebug
	}
	return &Logger{
		logLevel: level,
		out:      out,
	}
}

func (l *Logger) trace(format string, args ...any) {
	if l.logLevel > LevelTrace {
		return
	}
	fmt.Fprintf(l.out, "[TRACE] %s %s", l.times(), fmt.Sprintf(format, args...))
}

func (l *Logger) debug(format string, args ...any) {
	if l.logLevel > LevelDebug {
		return
	}
	fmt.Fprintf(l.out, "[TRACE] %s %s", l.times(), fmt.Sprintf(format, args...))
}

func (l *Logger) info(format string, args ...any) {
	if l.logLevel > LevelInfo {
		return
	}
	fmt.Fprintf(l.out, "[TRACE] %s %s", l.times(), fmt.Sprintf(format, args...))
}

func (l *Logger) warn(format string, args ...any) {
	if l.logLevel > LevelWarn {
		return
	}
	fmt.Fprintf(l.out, "[TRACE] %s %s", l.times(), fmt.Sprintf(format, args...))
}

func (l *Logger) error(format string, args ...any) {
	if l.logLevel > LevelError {
		return
	}
	fmt.Fprintf(l.out, "[TRACE] %s %s", l.times(), fmt.Sprintf(format, args...))
}

func (l *Logger) times() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

type LoggerKey struct{}

func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(LoggerKey{}).(*Logger)
	if !ok {
		return NewLogger("debug", os.Stdout)
	}

	return logger
}

func IntoContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, LoggerKey{}, logger)
}

func Trace(ctx context.Context, format string, args ...any) {
	FromContext(ctx).trace(format, args...)
}

func Debug(ctx context.Context, format string, args ...any) {
	FromContext(ctx).debug(format, args...)
}

func Info(ctx context.Context, format string, args ...any) {
	FromContext(ctx).info(format, args...)
}

func Warn(ctx context.Context, format string, args ...any) {
	FromContext(ctx).warn(format, args...)
}

func Error(ctx context.Context, format string, args ...any) {
	FromContext(ctx).error(format, args...)
}
