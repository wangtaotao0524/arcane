package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"github.com/ofkm/arcane-backend/internal/database"
	"github.com/ofkm/arcane-backend/internal/dto"
	"github.com/ofkm/arcane-backend/internal/models"
)

type UpdaterService struct {
	db                 *database.DB
	settingsService    *SettingsService
	dockerService      *DockerClientService
	stackService       *StackService
	imageUpdateService *ImageUpdateService
	registryService    *ContainerRegistryService
	eventService       *EventService
	imageService       *ImageService

	updatingContainers map[string]bool
	updatingStacks     map[string]bool
}

func NewUpdaterService(
	db *database.DB,
	settings *SettingsService,
	docker *DockerClientService,
	stacks *StackService,
	imageUpdates *ImageUpdateService,
	registries *ContainerRegistryService,
	events *EventService,
	imageSvc *ImageService,
) *UpdaterService {
	return &UpdaterService{
		db:                 db,
		settingsService:    settings,
		dockerService:      docker,
		stackService:       stacks,
		imageUpdateService: imageUpdates,
		registryService:    registries,
		eventService:       events,
		imageService:       imageSvc,
		updatingContainers: map[string]bool{},
		updatingStacks:     map[string]bool{},
	}
}

// ApplyPending pulls images and restarts impacted containers/stacks for all image_updates with has_update = true.
// - For tag updates: repo:latestVersion
// - For digest updates: keep tag ref (repo:tag), pulling refreshes digest; recreate with same ref.
func (s *UpdaterService) ApplyPending(ctx context.Context, dryRun bool) (*dto.UpdaterRunResult, error) {
	start := time.Now()
	out := &dto.UpdaterRunResult{Items: []dto.UpdaterItem{}}

	// enabled := s.settingsService.GetBoolSetting(ctx, "autoUpdate", false)
	// if !enabled {
	// 	out.Duration = time.Since(start).String()
	// 	return out, nil
	// }

	var records []models.ImageUpdateRecord
	if err := s.db.WithContext(ctx).Where("has_update = ?", true).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("query pending image updates: %w", err)
	}
	if len(records) == 0 {
		out.Duration = time.Since(start).String()
		return out, nil
	}

	// Only update images that are actually used by running resources
	usedImages, err := s.collectUsedImages(ctx)
	if err != nil {
		// Non-fatal: continue without the filter
		usedImages = map[string]struct{}{}
	}

	// Plan updates and capture OLD image digests before pull
	type updatePlan struct {
		oldRef string
		newRef string
		oldIDs []string // sha256:... image IDs that currently back oldRef
	}
	var plans []updatePlan

	for _, r := range records {
		if r.Repository == "" || r.Tag == "" {
			continue
		}
		oldRef := fmt.Sprintf("%s:%s", r.Repository, r.Tag)
		oldNorm := s.normalizeRef(oldRef)

		if len(usedImages) > 0 {
			if _, ok := usedImages[oldNorm]; !ok {
				continue
			}
		}

		newRef := oldRef
		if r.IsTagUpdate() && r.LatestVersion != nil && *r.LatestVersion != "" {
			newRef = fmt.Sprintf("%s:%s", r.Repository, *r.LatestVersion)
		}

		oldIDs, _ := s.resolveLocalImageIDsForRef(ctx, oldRef)
		plans = append(plans, updatePlan{oldRef: oldRef, newRef: newRef, oldIDs: oldIDs})
	}

	if len(plans) == 0 {
		out.Duration = time.Since(start).String()
		return out, nil
	}

	// Build maps for fast matching later
	oldRefToNewRef := map[string]string{}
	oldIDToNewRef := map[string]string{} // sha256 -> newRef
	for _, p := range plans {
		oldRefToNewRef[p.oldRef] = p.newRef
		for _, id := range p.oldIDs {
			oldIDToNewRef[id] = p.newRef
		}
	}

	// Pull images with ImageService (waits for completion)
	for _, p := range plans {
		item := dto.UpdaterItem{
			ResourceID:   p.oldRef,
			ResourceType: "image",
			ResourceName: p.oldRef,
			Status:       "checked",
			OldImages:    map[string]string{"main": p.oldRef},
			NewImages:    map[string]string{"main": p.newRef},
		}
		out.Checked++

		if dryRun {
			item.Status = "skipped"
			out.Skipped++
			out.Items = append(out.Items, item)
			_ = s.recordRun(ctx, item)
			continue
		}

		if err := s.imageService.PullImage(ctx, p.newRef, io.Discard, systemUser); err != nil {
			item.Status = "failed"
			item.Error = err.Error()
			out.Failed++
		} else {
			item.Status = "updated"
			item.UpdateApplied = true
			out.Updated++
		}
		out.Items = append(out.Items, item)
		_ = s.recordRun(ctx, item)
	}

	// Restart containers using old image IDs (skip stack-managed)
	if !dryRun && len(oldIDToNewRef) > 0 {
		results, err := s.restartContainersUsingOldIDs(ctx, oldIDToNewRef, oldRefToNewRef)
		if err != nil {
			log.Printf("container restarts had errors: %v", err)
		}
		for _, r := range results {
			item := dto.UpdaterItem{
				ResourceID:    r.ResourceID,
				ResourceType:  "container",
				ResourceName:  r.ResourceName,
				Status:        r.Status,
				Error:         r.Error,
				OldImages:     r.OldImages,
				NewImages:     r.NewImages,
				UpdateApplied: r.UpdateApplied,
			}
			out.Items = append(out.Items, item)
			out.Checked++
			switch {
			case r.UpdateApplied:
				out.Updated++
			case r.Error != "":
				out.Failed++
			default:
				out.Skipped++
			}
			_ = s.recordRun(ctx, item)
		}

		// Redeploy impacted stacks (only running ones)
		if err := s.redeployStacksUsingOldIDs(ctx, oldIDToNewRef, out); err != nil {
			log.Printf("stack redeploys had errors: %v", err)
		}
	}

	// Refresh image_updates by re-checking
	for oldRef := range oldRefToNewRef {
		if _, err := s.imageUpdateService.CheckImageUpdate(ctx, oldRef); err != nil {
			log.Printf("refresh update record failed for %s: %v", oldRef, err)
		}
	}

	out.Duration = time.Since(start).String()
	return out, nil
}

func (s *UpdaterService) GetStatus() map[string]any {
	containerIDs := make([]string, 0, len(s.updatingContainers))
	for id := range s.updatingContainers {
		containerIDs = append(containerIDs, id)
	}
	stackIDs := make([]string, 0, len(s.updatingStacks))
	for id := range s.updatingStacks {
		stackIDs = append(stackIDs, id)
	}

	return map[string]any{
		"updatingContainers": len(s.updatingContainers),
		"updatingStacks":     len(s.updatingStacks),
		"containerIds":       containerIDs,
		"stackIds":           stackIDs,
	}
}

func (s *UpdaterService) GetHistory(ctx context.Context, limit int) ([]models.AutoUpdateRecord, error) {
	var rec []models.AutoUpdateRecord
	q := s.db.WithContext(ctx).Order("start_time DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rec).Error; err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	return rec, nil
}

// --- internals ---

func (s *UpdaterService) updateContainer(ctx context.Context, cnt container.Summary, inspect container.InspectResponse, newRef string) error {
	dcli, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return fmt.Errorf("docker connect: %w", err)
	}
	defer dcli.Close()

	name := s.getContainerName(cnt)

	// stop
	if err := dcli.ContainerStop(ctx, cnt.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("stop: %w", err)
	}
	_ = s.eventService.LogContainerEvent(ctx, models.EventTypeContainerStop, cnt.ID, name, systemUser.ID, systemUser.Username, "0", models.JSON{"action": "updater_stop"})

	// remove
	if err := dcli.ContainerRemove(ctx, cnt.ID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	_ = s.eventService.LogContainerEvent(ctx, models.EventTypeContainerDelete, cnt.ID, name, systemUser.ID, systemUser.Username, "0", models.JSON{"action": "updater_delete"})

	// recreate with new image ref
	cfg := inspect.Config
	cfg.Image = newRef
	resp, err := dcli.ContainerCreate(ctx, cfg, inspect.HostConfig, &network.NetworkingConfig{EndpointsConfig: inspect.NetworkSettings.Networks}, nil, inspect.Name)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	_ = s.eventService.LogContainerEvent(ctx, models.EventTypeContainerCreate, resp.ID, name, systemUser.ID, systemUser.Username, "0", models.JSON{"action": "updater_create", "newImageId": resp.ID})

	if err := dcli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	_ = s.eventService.LogContainerEvent(ctx, models.EventTypeContainerStart, resp.ID, name, systemUser.ID, systemUser.Username, "0", models.JSON{"action": "updater_start"})

	return nil
}

// normalizeRef returns a canonical "registry/repository:tag" without digest.
// Examples:
// - "redis:latest" -> "docker.io/library/redis:latest"
// - "nginx@sha256:..." -> "docker.io/library/nginx:latest" (if no tag was present, defaults to latest)
func (s *UpdaterService) normalizeRef(ref string) string {
	ref = s.stripDigest(ref)

	// Split tag
	tag := "latest"
	if i := strings.LastIndex(ref, ":"); i != -1 && strings.LastIndex(ref, "/") < i {
		tag = ref[i+1:]
		ref = ref[:i]
	}

	parts := strings.Split(ref, "/")
	domain := ""
	switch {
	case len(parts) > 0 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") || parts[0] == "localhost"):
		domain = strings.ToLower(parts[0])
		parts = parts[1:]
	default:
		domain = "docker.io"
	}
	repo := strings.Join(parts, "/")
	if domain == "docker.io" && !strings.Contains(repo, "/") {
		repo = "library/" + repo
	}

	// Canonical docker.io domain
	switch domain {
	case "index.docker.io", "registry-1.docker.io":
		domain = "docker.io"
	}
	return strings.ToLower(domain + "/" + repo + ":" + tag)
}

func (s *UpdaterService) stripDigest(ref string) string {
	if i := strings.Index(ref, "@"); i != -1 {
		return ref[:i]
	}
	return ref
}

// collectUsedImages builds a set of normalized image refs used by running containers and running stacks.
func (s *UpdaterService) collectUsedImages(ctx context.Context) (map[string]struct{}, error) {
	out := map[string]struct{}{}

	// Running containers
	dcli, err := s.dockerService.CreateConnection(ctx)
	if err == nil {
		defer dcli.Close()
		list, err := dcli.ContainerList(ctx, container.ListOptions{All: false})
		if err == nil {
			for _, c := range list {
				inspect, err := dcli.ContainerInspect(ctx, c.ID)
				if err != nil {
					continue
				}
				for _, t := range s.getNormalizedTagsForContainer(ctx, dcli, inspect) {
					out[t] = struct{}{}
				}
			}
		}
	}

	// Running stacks (service definitions)
	stacks, err2 := s.stackService.ListStacks(ctx)
	if err2 == nil {
		for _, st := range stacks {
			if st.Status != models.StackStatusRunning {
				continue
			}
			srvs, err := s.stackService.GetStackServices(ctx, st.ID)
			if err != nil {
				continue
			}
			for _, svc := range srvs {
				if svc.Image == "" {
					continue
				}
				out[s.normalizeRef(svc.Image)] = struct{}{}
			}
		}
	}

	return out, nil
}

// Resolve a container's image to all normalized repo:tags that reference it.
// Falls back to the config image reference as a tag if present.
func (s *UpdaterService) getNormalizedTagsForContainer(ctx context.Context, dcli *client.Client, inspect container.InspectResponse) []string {
	seen := map[string]struct{}{}

	// Prefer tags from the image object (handles sha256 IDs)
	if dcli != nil {
		if ii, err := dcli.ImageInspect(ctx, inspect.Image); err == nil {
			for _, tag := range ii.RepoTags {
				if tag == "<none>:<none>" || tag == "" {
					continue
				}
				seen[s.normalizeRef(tag)] = struct{}{}
			}
		}
	}

	// Fallback to the configured image ref (may already be a tag)
	if inspect.Config != nil && inspect.Config.Image != "" {
		seen[s.normalizeRef(inspect.Config.Image)] = struct{}{}
	}

	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

func (s *UpdaterService) getContainerName(cnt container.Summary) string {
	if len(cnt.Names) > 0 {
		n := cnt.Names[0]
		if strings.HasPrefix(n, "/") {
			return n[1:]
		}
		return n
	}
	return cnt.ID[:12]
}

func (s *UpdaterService) recordRun(ctx context.Context, item dto.UpdaterItem) error {
	rec := &models.AutoUpdateRecord{
		ResourceID:      item.ResourceID,
		ResourceType:    item.ResourceType,
		ResourceName:    item.ResourceName,
		Status:          models.AutoUpdateStatus(item.Status),
		StartTime:       time.Now(),
		UpdateAvailable: item.Status == "updated" || item.Status == "update_available",
		UpdateApplied:   item.UpdateApplied,
	}

	if item.Error != "" {
		rec.Error = &item.Error
	}

	if len(item.OldImages) > 0 {
		old := make(models.JSON)
		for k, v := range item.OldImages {
			old[k] = v
		}
		rec.OldImageVersions = old
	}
	if len(item.NewImages) > 0 {
		newv := make(models.JSON)
		for k, v := range item.NewImages {
			newv[k] = v
		}
		rec.NewImageVersions = newv
	}

	end := time.Now()
	rec.EndTime = &end

	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *UpdaterService) isPartOfStack(labels map[string]string) bool {
	if labels == nil {
		return false
	}
	// Swarm stack label
	if _, ok := labels["com.docker.stack.namespace"]; ok {
		return true
	}
	// Compose project label
	if _, ok := labels["com.docker.compose.project"]; ok {
		return true
	}
	return false
}

// Resolve the local image ID(s) currently referenced by ref (repo:tag) before we pull.
// Returns IDs like "sha256:...".
func (s *UpdaterService) resolveLocalImageIDsForRef(ctx context.Context, ref string) ([]string, error) {
	dcli, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer dcli.Close()

	ids := []string{}
	if ii, err := dcli.ImageInspect(ctx, ref); err == nil && ii.ID != "" {
		ids = append(ids, ii.ID)
	}
	return ids, nil
}

// Containers: match by old image ID. Fallback to tag match if needed.
func (s *UpdaterService) restartContainersUsingOldIDs(ctx context.Context, oldIDToNewRef map[string]string, oldRefToNewRef map[string]string) ([]dto.AutoUpdateResourceResult, error) {
	dcli, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("docker connect: %w", err)
	}
	defer dcli.Close()

	list, err := dcli.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	var results []dto.AutoUpdateResourceResult
	for _, c := range list {
		if s.isPartOfStack(c.Labels) {
			continue
		}
		inspect, err := dcli.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}

		var (
			newRef string
			match  string
		)

		// Primary: match by digest (image ID)
		if nr, ok := oldIDToNewRef[inspect.Image]; ok {
			newRef = nr
			match = inspect.Image
		} else {
			// Fallback: resolve tags and match by tag
			updatedNorm := map[string]string{}
			for oldRef, nr := range oldRefToNewRef {
				updatedNorm[s.normalizeRef(oldRef)] = nr
			}
			for _, t := range s.getNormalizedTagsForContainer(ctx, dcli, inspect) {
				if nr, ok := updatedNorm[t]; ok {
					newRef = nr
					match = t
					break
				}
			}
			if newRef == "" {
				continue
			}
		}

		name := s.getContainerName(c)
		res := dto.AutoUpdateResourceResult{
			ResourceID:   c.ID,
			ResourceName: name,
			ResourceType: "container",
			Status:       "checked",
			OldImages:    map[string]string{"main": match},
			NewImages:    map[string]string{"main": s.normalizeRef(newRef)},
		}

		if err := s.updateContainer(ctx, c, inspect, newRef); err != nil {
			res.Status = "failed"
			res.Error = err.Error()
		} else {
			res.Status = "updated"
			res.UpdateAvailable = true
			res.UpdateApplied = true
		}
		results = append(results, res)
	}
	return results, nil
}

// Stacks: impacted if any container in the stack runs an image ID in oldIDToNewRef.
func (s *UpdaterService) redeployStacksUsingOldIDs(ctx context.Context, oldIDToNewRef map[string]string, out *dto.UpdaterRunResult) error {
	stacks, err := s.stackService.ListStacks(ctx)
	if err != nil {
		return fmt.Errorf("list stacks: %w", err)
	}

	dcli, derr := s.dockerService.CreateConnection(ctx)
	if derr != nil {
		return fmt.Errorf("docker connect: %w", derr)
	}
	defer dcli.Close()

	for _, st := range stacks {
		// Only redeploy stacks that are currently running
		if st.Status != models.StackStatusRunning {
			out.Items = append(out.Items, dto.UpdaterItem{
				ResourceID:   st.ID,
				ResourceType: "stack",
				ResourceName: st.Name,
				Status:       "skipped",
			})
			out.Skipped++
			out.Checked++
			continue
		}

		containers, lerr := s.getStackContainers(ctx, dcli, st.Name)
		if lerr != nil {
			log.Printf("list stack containers %s: %v", st.Name, lerr)
			continue
		}

		impacted := false
		for _, c := range containers {
			inspect, ierr := dcli.ContainerInspect(ctx, c.ID)
			if ierr != nil {
				continue
			}
			if _, ok := oldIDToNewRef[inspect.Image]; ok {
				impacted = true
				break
			}
		}

		if !impacted {
			out.Items = append(out.Items, dto.UpdaterItem{
				ResourceID:   st.ID,
				ResourceType: "stack",
				ResourceName: st.Name,
				Status:       "skipped",
			})
			out.Skipped++
			out.Checked++
			continue
		}

		item := dto.UpdaterItem{
			ResourceID:   st.ID,
			ResourceType: "stack",
			ResourceName: st.Name,
			Status:       "checked",
		}
		out.Checked++

		// Pull, down, deploy to avoid conflicts and ensure fresh images
		if err := s.stackService.PullStackImages(ctx, st.ID, io.Discard); err != nil {
			log.Printf("stack pull warning: %v", err)
		}
		if err := s.stackService.RedeployStack(ctx, st.ID, nil, nil, systemUser); err != nil {
			item.Status = "failed"
			item.Error = err.Error()
			out.Failed++
		} else {
			item.Status = "updated"
			item.UpdateApplied = true
			out.Updated++
		}
		out.Items = append(out.Items, item)
		_ = s.recordRun(ctx, item)
	}
	return nil
}

// getStackContainers returns the union of containers matching compose or swarm stack labels (OR semantics).
func (s *UpdaterService) getStackContainers(ctx context.Context, dcli *client.Client, stackName string) ([]container.Summary, error) {
	byID := map[string]container.Summary{}

	// Compose label
	f1 := filters.NewArgs()
	f1.Add("label", "com.docker.compose.project="+stackName)
	cs1, err1 := dcli.ContainerList(ctx, container.ListOptions{All: true, Filters: f1})
	if err1 == nil {
		for _, c := range cs1 {
			byID[c.ID] = c
		}
	}

	// Swarm namespace label
	f2 := filters.NewArgs()
	f2.Add("label", "com.docker.stack.namespace="+stackName)
	cs2, err2 := dcli.ContainerList(ctx, container.ListOptions{All: true, Filters: f2})
	if err2 == nil {
		for _, c := range cs2 {
			byID[c.ID] = c
		}
	}

	// If both failed, return the last error
	if err1 != nil && err2 != nil {
		return nil, err2
	}

	out := make([]container.Summary, 0, len(byID))
	for _, c := range byID {
		out = append(out, c)
	}
	return out, nil
}
