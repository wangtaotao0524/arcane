package job

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ofkm/arcane-backend/internal/services"
	"github.com/ofkm/arcane-backend/internal/utils"
)

type FilesystemWatcherJob struct {
	stackService    *services.StackService
	settingsService *services.SettingsService
	watcher         *utils.FilesystemWatcher
}

func NewFilesystemWatcherJob(
	stackService *services.StackService,
	settingsService *services.SettingsService,
) *FilesystemWatcherJob {
	return &FilesystemWatcherJob{
		stackService:    stackService,
		settingsService: settingsService,
	}
}

func RegisterFilesystemWatcherJob(ctx context.Context, scheduler *Scheduler, stackService *services.StackService, settingsService *services.SettingsService) error {
	job := NewFilesystemWatcherJob(stackService, settingsService)

	// Start the filesystem watcher as a background task
	go func() {
		if err := job.Start(ctx); err != nil {
			slog.ErrorContext(ctx, "Filesystem watcher failed", "error", err)
		}
	}()

	slog.InfoContext(ctx, "Filesystem watcher job registered")
	return nil
}

func (j *FilesystemWatcherJob) Start(ctx context.Context) error {
	// Get stacks directory from settings
	stacksDir, err := j.getStacksDirectory(ctx)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if _, err := os.Stat(stacksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(stacksDir, 0755); err != nil {
			return err
		}
		slog.InfoContext(ctx, "Created stacks directory", "path", stacksDir)
	}

	// Create filesystem watcher
	watcher, err := utils.NewFilesystemWatcher(stacksDir, utils.WatcherOptions{
		Debounce: 3 * time.Second, // Wait 3 seconds after last change before syncing
		OnChange: j.handleFilesystemChange,
	})
	if err != nil {
		return err
	}

	j.watcher = watcher

	// Start watching
	if err := j.watcher.Start(ctx); err != nil {
		return err
	}

	slog.InfoContext(ctx, "Filesystem watcher started for stacks directory",
		"path", stacksDir)

	// Keep the watcher running until context is cancelled
	<-ctx.Done()

	return j.Stop()
}

func (j *FilesystemWatcherJob) Stop() error {
	if j.watcher != nil {
		return j.watcher.Stop()
	}
	return nil
}

func (j *FilesystemWatcherJob) handleFilesystemChange(ctx context.Context) {
	slog.InfoContext(ctx, "Filesystem change detected, syncing stacks")

	if err := j.stackService.SyncAllStacksFromFilesystem(ctx); err != nil {
		slog.ErrorContext(ctx, "Failed to sync stacks after filesystem change",
			"error", err)
	} else {
		slog.InfoContext(ctx, "Stack sync completed after filesystem change")
	}
}

func (j *FilesystemWatcherJob) getStacksDirectory(ctx context.Context) (string, error) {
	settings, err := j.settingsService.GetSettings(ctx)
	if err != nil {
		return "data/projects", err
	}
	return settings.StacksDirectory.Value, nil
}
