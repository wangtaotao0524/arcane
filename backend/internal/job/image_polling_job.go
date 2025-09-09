package job

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/ofkm/arcane-backend/internal/services"
)

type ImagePollingJob struct {
	imageUpdateService *services.ImageUpdateService
	settingsService    *services.SettingsService
}

func NewImagePollingJob(imageUpdateService *services.ImageUpdateService, settingsService *services.SettingsService) *ImagePollingJob {
	return &ImagePollingJob{
		imageUpdateService: imageUpdateService,
		settingsService:    settingsService,
	}
}

func RegisterImagePollingJob(ctx context.Context, scheduler *Scheduler, imageUpdateService *services.ImageUpdateService, settingsService *services.SettingsService) error {
	pollingEnabled := settingsService.GetBoolSetting(ctx, "pollingEnabled", true)
	pollingInterval := settingsService.GetIntSetting(ctx, "pollingInterval", 60)

	if !pollingEnabled {
		slog.InfoContext(ctx, "polling disabled; job not registered")
		return nil
	}

	interval := time.Duration(pollingInterval) * time.Minute
	if interval < 5*time.Minute {
		slog.WarnContext(ctx, "polling interval too low; using default",
			"requested_seconds", pollingInterval,
			"effective_interval", "60m")
		interval = 60 * time.Minute
	}

	slog.InfoContext(ctx, "registering image polling job", "interval", interval.String())

	job := NewImagePollingJob(imageUpdateService, settingsService)
	jobDefinition := gocron.DurationJob(interval)

	return scheduler.RegisterJob(
		ctx,
		"image-polling",
		jobDefinition,
		job.Execute,
		false,
	)
}

func (j *ImagePollingJob) Execute(ctx context.Context) error {
	slog.InfoContext(ctx, "image scan run started")

	results, err := j.imageUpdateService.CheckAllImages(ctx, 0)
	if err != nil {
		slog.ErrorContext(ctx, "image scan failed", "err", err)
		return err
	}

	total := len(results)
	updates := 0
	errors := 0
	for _, r := range results {
		if r == nil {
			continue
		}
		if r.Error != "" {
			errors++
			continue
		}
		if r.HasUpdate {
			updates++
		}
	}

	slog.InfoContext(ctx, "image scan run completed",
		"checked", total,
		"updates", updates,
		"errors", errors,
	)

	return nil
}

func UpdateImagePollingJobSchedule(ctx context.Context, scheduler *Scheduler, imageUpdateService *services.ImageUpdateService, settingsService *services.SettingsService) error {
	pollingEnabled := settingsService.GetBoolSetting(ctx, "pollingEnabled", true)
	pollingInterval := settingsService.GetIntSetting(ctx, "pollingInterval", 300)

	if !pollingEnabled {
		slog.InfoContext(ctx, "polling job disabled; job will be skipped")
		return nil
	}

	interval := time.Duration(pollingInterval) * time.Second
	slog.InfoContext(ctx, "polling settings changed", "interval", interval.String())

	return nil
}
