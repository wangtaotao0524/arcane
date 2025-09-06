package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"

	"github.com/ofkm/arcane-backend/internal/config"
	"github.com/ofkm/arcane-backend/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupLogger(cfg *config.Config) {
	var lvl slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	lv := new(slog.LevelVar)
	lv.Set(lvl)

	var h slog.Handler
	if cfg.LogJson {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	} else {
		h = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      lv,
			TimeFormat: "Jan 02 15:04:05.000",
		})
	}

	slog.SetDefault(slog.New(h))
}

func BuildGormLogger(cfg *config.Config) logger.Interface {
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		return newGormSlogLogger(logger.Info, 200*time.Millisecond, true)
	case "info":
		return newGormSlogLogger(logger.Warn, 200*time.Millisecond, true)
	case "warn", "warning":
		return newGormSlogLogger(logger.Error, 200*time.Millisecond, true)
	case "error":
		return newGormSlogLogger(logger.Error, 200*time.Millisecond, true)
	default:
		return newGormSlogLogger(logger.Warn, 200*time.Millisecond, true)
	}
}

func ConfigureGormLogger(cfg *config.Config) {
	database.SetGormLogger(BuildGormLogger(cfg))
}

type gormSlogLogger struct {
	level                     logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

func newGormSlogLogger(level logger.LogLevel, slow time.Duration, ignoreNotFound bool) logger.Interface {
	return &gormSlogLogger{
		level:                     level,
		slowThreshold:             slow,
		ignoreRecordNotFoundError: ignoreNotFound,
	}
}

func (l *gormSlogLogger) LogMode(level logger.LogLevel) logger.Interface {
	nl := *l
	nl.level = level
	return &nl
}

func (l *gormSlogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level > logger.Info {
		return
	}
	slog.InfoContext(ctx, msg, slog.Any("data", data))
}

func (l *gormSlogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level > logger.Warn {
		return
	}
	slog.WarnContext(ctx, msg, slog.Any("data", data))
}

func (l *gormSlogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
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
	oneLine := strings.Join(strings.Fields(s), " ")
	if len(oneLine) > maxSQLLen {
		return oneLine[:maxSQLLen] + "…"
	}
	return oneLine
}

func (l *gormSlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
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

	anyAttrs := make([]any, 0, len(attrs)+1)
	for _, a := range attrs {
		anyAttrs = append(anyAttrs, a)
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
