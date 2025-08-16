package database

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// slogGormLogger implements gorm/logger.Interface using Go's slog for JSON logs.
type slogGormLogger struct {
	level                     logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

func newSlogGormLogger(level logger.LogLevel, slow time.Duration, ignoreNotFound bool) *slogGormLogger {
	return &slogGormLogger{level: level, slowThreshold: slow, ignoreRecordNotFoundError: ignoreNotFound}
}

func (l *slogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	nl := *l
	nl.level = level
	return &nl
}

func (l *slogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level > logger.Info {
		return
	}
	slog.InfoContext(ctx, msg, slog.Any("data", data))
}

func (l *slogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level > logger.Warn {
		return
	}
	slog.WarnContext(ctx, msg, slog.Any("data", data))
}

func (l *slogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level > logger.Error {
		return
	}
	slog.ErrorContext(ctx, msg, slog.Any("data", data))
}

const maxSQLLen = 500

func sanitizeSQL(s string) string {
	if s == "" {
		return s
	}
	// Trim and convert any whitespace/newlines/tabs to single spaces
	oneLine := strings.Join(strings.Fields(s), " ")
	if len(oneLine) > maxSQLLen {
		return oneLine[:maxSQLLen] + "…"
	}
	return oneLine
}

func (l *slogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	sql = sanitizeSQL(sql)

	attrs := []slog.Attr{
		slog.Int64("elapsed_ms", elapsed.Milliseconds()),
	}
	if rows >= 0 {
		attrs = append(attrs, slog.Int64("rows", rows))
	}
	if sql != "" {
		attrs = append(attrs, slog.String("sql", sql))
	}

	// convert []slog.Attr to []any for slog.*Context variadic parameter
	anyAttrs := make([]any, len(attrs))
	for i, a := range attrs {
		anyAttrs[i] = a
	}

	if err != nil && l.level <= logger.Error {
		if !l.ignoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound) {
			anyAttrs = append(anyAttrs, slog.String("error", err.Error()))
			slog.ErrorContext(ctx, "gorm.error", anyAttrs...)
		}
		return
	}

	if l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level <= logger.Warn {
		slog.WarnContext(ctx, "gorm.slow_query", anyAttrs...)
		return
	}

	if l.level <= logger.Info {
		slog.InfoContext(ctx, "gorm.query", anyAttrs...)
	}
}
