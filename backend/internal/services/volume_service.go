package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/ofkm/arcane-backend/internal/database"
	"github.com/ofkm/arcane-backend/internal/models"
	"github.com/ofkm/arcane-backend/internal/utils"
)

type VolumeService struct {
	db            *database.DB
	dockerService *DockerClientService
	eventService  *EventService
}

func NewVolumeService(db *database.DB, dockerService *DockerClientService, eventService *EventService) *VolumeService {
	return &VolumeService{
		db:            db,
		dockerService: dockerService,
		eventService:  eventService,
	}
}

func (s *VolumeService) buildVolumeUsageMap(ctx context.Context, dockerClient *client.Client) (map[string]bool, error) {
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	usage := make(map[string]bool)
	for _, c := range containers {
		info, err := dockerClient.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}
		for _, m := range info.Mounts {
			if m.Type == "volume" && m.Name != "" {
				usage[m.Name] = true
			}
		}
	}
	return usage, nil
}

func (s *VolumeService) containersUsingVolume(ctx context.Context, dockerClient *client.Client, name string) (bool, []string, error) {
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, nil, fmt.Errorf("failed to list containers: %w", err)
	}

	inUse := false
	var using []string
	for _, c := range containers {
		info, err := dockerClient.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}
		for _, m := range info.Mounts {
			if m.Type == "volume" && m.Name == name {
				inUse = true
				using = append(using, c.ID)
				break
			}
		}
	}
	return inUse, using, nil
}

func toVolumeMap(v volume.Volume, inUse bool) map[string]interface{} {
	return map[string]interface{}{
		"Name":       v.Name,
		"Driver":     v.Driver,
		"Mountpoint": v.Mountpoint,
		"Scope":      v.Scope,
		"Options":    v.Options,
		"Labels":     v.Labels,
		"CreatedAt":  v.CreatedAt,
		"InUse":      inUse,
	}
}

func (s *VolumeService) ListVolumes(ctx context.Context) ([]volume.Volume, error) {
	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	volumes, err := dockerClient.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker volumes: %w", err)
	}

	vols := make([]volume.Volume, len(volumes.Volumes))
	for i, v := range volumes.Volumes {
		if v != nil {
			vols[i] = *v
		}
	}
	return vols, nil
}

func (s *VolumeService) GetVolumeByName(ctx context.Context, name string) (*volume.Volume, error) {
	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	vol, err := dockerClient.VolumeInspect(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("volume not found: %w", err)
	}

	return &vol, nil
}

func (s *VolumeService) CreateVolume(ctx context.Context, options volume.CreateOptions, user models.User) (*volume.Volume, error) {
	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	vol, err := dockerClient.VolumeCreate(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	if s.db != nil {
		dbVolume := &models.Volume{
			BaseModel:  models.BaseModel{CreatedAt: time.Now()},
			Name:       vol.Name,
			Driver:     vol.Driver,
			Mountpoint: vol.Mountpoint,
			Scope:      "local",
		}
		s.db.WithContext(ctx).Create(dbVolume)
	}

	// Log volume creation event
	metadata := models.JSON{
		"action": "create",
		"driver": vol.Driver,
		"name":   vol.Name,
	}
	if logErr := s.eventService.LogVolumeEvent(ctx, models.EventTypeVolumeCreate, vol.Name, vol.Name, user.ID, user.Username, "0", metadata); logErr != nil {
		fmt.Printf("Could not log volume creation action: %s\n", logErr)
	}

	return &vol, nil
}

func (s *VolumeService) DeleteVolume(ctx context.Context, name string, force bool, user models.User) error {
	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	if err := dockerClient.VolumeRemove(ctx, name, force); err != nil {
		return fmt.Errorf("failed to remove volume: %w", err)
	}

	if s.db != nil {
		s.db.WithContext(ctx).Delete(&models.Volume{}, "name = ?", name)
	}

	// Log volume deletion event
	metadata := models.JSON{
		"action": "delete",
		"name":   name,
		"force":  force,
	}
	if logErr := s.eventService.LogVolumeEvent(ctx, models.EventTypeVolumeDelete, name, name, user.ID, user.Username, "0", metadata); logErr != nil {
		fmt.Printf("Could not log volume deletion action: %s\n", logErr)
	}

	return nil
}

func (s *VolumeService) PruneVolumes(ctx context.Context) (*volume.PruneReport, error) {
	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	filterArgs := filters.NewArgs()

	report, err := dockerClient.VolumesPrune(ctx, filterArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to prune volumes: %w", err)
	}

	// Log volume prune event (system user)
	metadata := models.JSON{
		"action":         "prune",
		"volumesDeleted": len(report.VolumesDeleted),
		"spaceReclaimed": report.SpaceReclaimed,
	}
	if logErr := s.eventService.LogVolumeEvent(ctx, models.EventTypeVolumeDelete, "", "bulk_prune", systemUser.ID, systemUser.Username, "0", metadata); logErr != nil {
		fmt.Printf("Could not log volume prune action: %s\n", logErr)
	}

	return &report, nil
}

func (s *VolumeService) GetVolumesByDriver(ctx context.Context, driver string) ([]volume.Volume, error) {
	volumes, err := s.ListVolumes(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []volume.Volume
	for _, vol := range volumes {
		if vol.Driver == driver {
			filtered = append(filtered, vol)
		}
	}

	return filtered, nil
}

func (s *VolumeService) GetVolumeUsage(ctx context.Context, name string) (bool, []string, error) {
	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	if _, err := dockerClient.VolumeInspect(ctx, name); err != nil {
		return false, nil, fmt.Errorf("volume not found: %w", err)
	}

	inUse, usingContainers, err := s.containersUsingVolume(ctx, dockerClient, name)
	if err != nil {
		return false, nil, err
	}
	return inUse, usingContainers, nil
}

//nolint:gocognit
func (s *VolumeService) ListVolumesPaginated(ctx context.Context, req utils.SortedPaginationRequest) ([]map[string]interface{}, utils.PaginationResponse, error) {
	volumes, err := s.ListVolumes(ctx)
	if err != nil {
		return nil, utils.PaginationResponse{}, fmt.Errorf("failed to list Docker volumes: %w", err)
	}

	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, utils.PaginationResponse{}, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	volumeUsageMap, err := s.buildVolumeUsageMap(ctx, dockerClient)
	if err != nil {
		return nil, utils.PaginationResponse{}, err
	}

	var result []map[string]interface{}
	for _, v := range volumes {
		result = append(result, toVolumeMap(v, volumeUsageMap[v.Name]))
	}

	if req.Search != "" {
		filtered := make([]map[string]interface{}, 0, len(result))
		searchLower := strings.ToLower(req.Search)
		for _, vol := range result {
			if name, ok := vol["Name"].(string); ok && strings.Contains(strings.ToLower(name), searchLower) {
				filtered = append(filtered, vol)
				continue
			}
			if driver, ok := vol["Driver"].(string); ok && strings.Contains(strings.ToLower(driver), searchLower) {
				filtered = append(filtered, vol)
				continue
			}
		}
		result = filtered
	}

	totalItems := len(result)

	if req.Sort.Column != "" {
		utils.SortSliceByField(result, req.Sort.Column, req.Sort.Direction)
	}

	startIdx := (req.Pagination.Page - 1) * req.Pagination.Limit
	endIdx := startIdx + req.Pagination.Limit

	if startIdx > len(result) {
		startIdx = len(result)
	}
	if endIdx > len(result) {
		endIdx = len(result)
	}

	if startIdx < endIdx {
		result = result[startIdx:endIdx]
	} else {
		result = []map[string]interface{}{}
	}

	totalPages := (totalItems + req.Pagination.Limit - 1) / req.Pagination.Limit
	pagination := utils.PaginationResponse{
		TotalPages:   int64(totalPages),
		TotalItems:   int64(totalItems),
		CurrentPage:  req.Pagination.Page,
		ItemsPerPage: req.Pagination.Limit,
	}

	return result, pagination, nil
}

func (s *VolumeService) ListVolumesWithUsage(ctx context.Context) ([]map[string]interface{}, error) {
	volumes, err := s.ListVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker volumes: %w", err)
	}

	dockerClient, err := s.dockerService.CreateConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer dockerClient.Close()

	volumeUsageMap, err := s.buildVolumeUsageMap(ctx, dockerClient)
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, v := range volumes {
		result = append(result, toVolumeMap(v, volumeUsageMap[v.Name]))
	}

	return result, nil
}
