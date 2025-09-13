package job

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/ofkm/arcane-backend/internal/services"
)

type AutoUpdateJob struct {
	updaterService  *services.UpdaterService
	settingsService *services.SettingsService
}

func NewAutoUpdateJob(updaterService *services.UpdaterService, settingsService *services.SettingsService) *AutoUpdateJob {
	return &AutoUpdateJob{
		updaterService:  updaterService,
		settingsService: settingsService,
	}
}

func RegisterAutoUpdateJob(ctx context.Context, scheduler *Scheduler, updaterService *services.UpdaterService, settingsService *services.SettingsService) error {
	autoUpdateEnabled := settingsService.GetBoolSetting(ctx, "autoUpdate", false)
	autoUpdateInterval := settingsService.GetIntSetting(ctx, "autoUpdateInterval", 1440)

	if !autoUpdateEnabled {
		slog.InfoContext(ctx, "auto-update disabled; job not registered")
		return nil
	}

	interval := time.Duration(autoUpdateInterval) * time.Minute
	if interval < 5*time.Minute {
		slog.WarnContext(ctx, "auto-update interval too low; using default",
			"requested_seconds", autoUpdateInterval,
			"effective_interval", "60m")
		interval = 60 * time.Minute
	}

	slog.InfoContext(ctx, "registering auto-update job", "interval", interval.String())

	job := NewAutoUpdateJob(updaterService, settingsService)
	jobDefinition := gocron.DurationJob(interval)

	return scheduler.RegisterJob(
		ctx,
		"auto-update",
		jobDefinition,
		job.Execute,
		false,
	)
}

func (j *AutoUpdateJob) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "auto-update run started")

	enabled := j.settingsService.GetBoolSetting(ctx, "autoUpdate", false)
	if !enabled {
		slog.InfoContext(ctx, "auto-update disabled; skipping run")
		return nil
	}

	result, err := j.updaterService.ApplyPending(ctx, false)
	if err != nil {
		slog.ErrorContext(ctx, "auto-update run failed", "err", err)
		return err
	}

	slog.InfoContext(ctx, "auto-update run completed",
		"checked", result.Checked,
		"updated", result.Updated,
		"skipped", result.Skipped,
		"failed", result.Failed,
	)

	return nil
}

func UpdateAutoUpdateJobSchedule(ctx context.Context, scheduler *Scheduler, updaterService *services.UpdaterService, settingsService *services.SettingsService) error {
	autoUpdateEnabled := settingsService.GetBoolSetting(ctx, "autoUpdate", false)
	autoUpdateInterval := settingsService.GetIntSetting(ctx, "autoUpdateInterval", 300)

	if !autoUpdateEnabled {
		slog.InfoContext(ctx, "auto-update disabled; job will be skipped")
		return nil
	}

	interval := time.Duration(autoUpdateInterval) * time.Second
	slog.InfoContext(ctx, "auto-update settings changed", "interval", interval.String())

	return nil
}
