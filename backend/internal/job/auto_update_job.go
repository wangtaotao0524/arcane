package job

import (
	"context"
	"log"
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
	autoUpdateInterval := settingsService.GetIntSetting(ctx, "autoUpdateInterval", 300)

	if !autoUpdateEnabled {
		log.Println("Auto-update is disabled, not registering auto-update job")
		return nil
	}

	interval := time.Duration(autoUpdateInterval) * time.Second
	if interval < 5*time.Minute {
		interval = 60 * time.Minute
		log.Printf("Auto-update interval too low, using default 60 minutes")
	}

	log.Printf("Registering auto-update job with %v interval", interval)

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
	log.Println("=== Starting scheduled updater run ===")

	enabled := j.settingsService.GetBoolSetting(ctx, "autoUpdate", false)
	if !enabled {
		log.Println("Auto-update disabled, skipping")
		return nil
	}

	result, err := j.updaterService.ApplyPending(ctx, false)
	if err != nil {
		log.Printf("Updater error: %v", err)
		return err
	}

	log.Printf("Updater completed: checked=%d updated=%d skipped=%d failed=%d",
		result.Checked, result.Updated, result.Skipped, result.Failed)
	return nil
}

func UpdateAutoUpdateJobSchedule(ctx context.Context, scheduler *Scheduler, updaterService *services.UpdaterService, settingsService *services.SettingsService) error {
	autoUpdateEnabled := settingsService.GetBoolSetting(ctx, "autoUpdate", false)
	autoUpdateInterval := settingsService.GetIntSetting(ctx, "autoUpdateInterval", 300)

	if !autoUpdateEnabled {
		log.Println("Auto-update disabled, job will be skipped")
		return nil
	}

	interval := time.Duration(autoUpdateInterval) * time.Second
	log.Printf("Auto-update settings changed - new interval: %v", interval)
	return nil
}
