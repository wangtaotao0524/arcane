package job

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/ofkm/arcane-backend/internal/services"
)

const StackSyncJobName = "StackSync"

func RegisterStackSyncJob(
	ctx context.Context,
	scheduler *Scheduler,
	stackService *services.StackService,
) error {
	slog.InfoContext(ctx, "Registering stack sync job", "jobName", StackSyncJobName)

	taskFunc := func(jobCtx context.Context) error {
		slog.InfoContext(jobCtx, "Running stack sync job", "jobName", StackSyncJobName)

		// Sync all stacks with filesystem (this handles both discovery and updates)
		if err := stackService.SyncAllStacksFromFilesystem(jobCtx); err != nil {
			slog.ErrorContext(jobCtx, "Failed to sync stacks with filesystem",
				"jobName", StackSyncJobName,
				slog.Any("error", err))
			return err
		}

		slog.InfoContext(jobCtx, "Stack sync job completed", "jobName", StackSyncJobName)
		return nil
	}

	jobDefinition := gocron.DurationJob(2 * time.Minute)

	err := scheduler.RegisterJob(
		ctx,
		StackSyncJobName,
		jobDefinition,
		taskFunc,
		true, // Run immediately on startup
	)

	if err != nil {
		return fmt.Errorf("failed to register stack sync job %q: %w", StackSyncJobName, err)
	}

	slog.InfoContext(ctx, "Stack sync job registered successfully", "jobName", StackSyncJobName)
	return nil
}
