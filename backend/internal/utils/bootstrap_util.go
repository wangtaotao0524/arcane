package utils

import (
	"context"
	"log/slog"

	"github.com/ofkm/arcane-backend/internal/config"
)

func LoadAgentToken(ctx context.Context, cfg *config.Config, getSettingFunc func(context.Context, string, string) string) {
	if cfg.AgentMode && cfg.AgentToken == "" {
		if tok := getSettingFunc(ctx, "agentToken", ""); tok != "" {
			cfg.AgentToken = tok
			slog.InfoContext(ctx, "Loaded agent token from database")
		}
	}
}

func EnsureEncryptionKey(ctx context.Context, cfg *config.Config, ensureKeyFunc func(context.Context) (string, error)) {
	if cfg.AgentMode || cfg.Environment != "production" {
		key, err := ensureKeyFunc(ctx)
		if err != nil {
			slog.WarnContext(ctx, "Failed to ensure encryption key; falling back to derived behavior",
				slog.String("error", err.Error()))
			return
		}
		cfg.EncryptionKey = key
	}
}

type SettingsManager interface {
	PersistEnvSettingsIfMissing(ctx context.Context) error
	SetBoolSetting(ctx context.Context, key string, value bool) error
	EnsureDefaultSettings(ctx context.Context) error
}

func InitializeDefaultSettings(ctx context.Context, cfg *config.Config, settingsMgr SettingsManager) {
	slog.InfoContext(ctx, "Ensuring default settings are initialized")

	if err := settingsMgr.EnsureDefaultSettings(ctx); err != nil {
		slog.WarnContext(ctx, "Failed to initialize default settings", slog.String("error", err.Error()))
	} else {
		slog.InfoContext(ctx, "Default settings initialized successfully")
	}

	// Mark onboarding as completed for all installs (onboarding is replaced with first-login password change)
	if err := settingsMgr.SetBoolSetting(ctx, "onboardingCompleted", true); err != nil {
		slog.WarnContext(ctx, "Failed to mark onboarding as completed", slog.String("error", err.Error()))
	} else {
		slog.InfoContext(ctx, "Onboarding marked as completed")
	}

	if cfg.AgentMode || cfg.UIConfigurationDisabled {
		if err := settingsMgr.PersistEnvSettingsIfMissing(ctx); err != nil {
			slog.WarnContext(ctx, "Failed to persist env-driven settings", slog.String("error", err.Error()))
		} else {
			slog.DebugContext(ctx, "Persisted env-driven settings if missing")
		}
	}
}

func TestDockerConnection(ctx context.Context, testFunc func(context.Context) error) {
	if err := testFunc(ctx); err != nil {
		slog.WarnContext(ctx, "Docker connection failed during init, local Docker features may be unavailable",
			slog.String("error", err.Error()))
	}
}

func InitializeNonAgentFeatures(ctx context.Context, cfg *config.Config, createAdminFunc func() error, syncOidcFunc func(context.Context) error) {
	if cfg.AgentMode {
		return
	}

	if err := createAdminFunc(); err != nil {
		slog.WarnContext(ctx, "Failed to create default admin user",
			slog.String("error", err.Error()))
	}

	if cfg.OidcEnabled {
		if err := syncOidcFunc(ctx); err != nil {
			slog.WarnContext(ctx, "Failed to sync OIDC environment variables to database",
				slog.String("error", err.Error()))
		}
	}
}
