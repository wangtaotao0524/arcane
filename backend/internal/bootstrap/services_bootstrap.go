package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ofkm/arcane-backend/internal/config"
	"github.com/ofkm/arcane-backend/internal/database"
	"github.com/ofkm/arcane-backend/internal/services"
	"github.com/ofkm/arcane-backend/resources"
)

type Services struct {
	AppImages         *services.ApplicationImagesService
	User              *services.UserService
	Project           *services.ProjectService
	Environment       *services.EnvironmentService
	Settings          *services.SettingsService
	SettingsSearch    *services.SettingsSearchService
	CustomizeSearch   *services.CustomizeSearchService
	Container         *services.ContainerService
	Image             *services.ImageService
	Volume            *services.VolumeService
	Network           *services.NetworkService
	ImageUpdate       *services.ImageUpdateService
	Auth              *services.AuthService
	Oidc              *services.OidcService
	Docker            *services.DockerClientService
	Template          *services.TemplateService
	ContainerRegistry *services.ContainerRegistryService
	System            *services.SystemService
	Updater           *services.UpdaterService
	Event             *services.EventService
	Version           *services.VersionService
	SSH               *services.SSHService
}

func initializeServices(ctx context.Context, db *database.DB, cfg *config.Config, httpClient *http.Client) (svcs *Services, dockerSrvice *services.DockerClientService, err error) {
	svcs = &Services{}

	svcs.Event = services.NewEventService(db)
	svcs.Settings, err = services.NewSettingsService(ctx, db)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to settings service: %w", err)
	}
	svcs.SettingsSearch = services.NewSettingsSearchService()
	svcs.CustomizeSearch = services.NewCustomizeSearchService()
	svcs.AppImages = services.NewApplicationImagesService(resources.FS, svcs.Settings)
	dockerClient := services.NewDockerClientService(db, cfg)
	svcs.Docker = dockerClient
	svcs.User = services.NewUserService(db)
	svcs.ContainerRegistry = services.NewContainerRegistryService(db)
	svcs.ImageUpdate = services.NewImageUpdateService(db, svcs.Settings, svcs.ContainerRegistry, svcs.Docker, svcs.Event)
	svcs.Image = services.NewImageService(db, svcs.Docker, svcs.ContainerRegistry, svcs.ImageUpdate, svcs.Event)
	svcs.Project = services.NewProjectService(db, svcs.Settings, svcs.Event, svcs.Image)
	svcs.Environment = services.NewEnvironmentService(db, httpClient)
	svcs.Container = services.NewContainerService(db, svcs.Event, svcs.Docker)
	svcs.Volume = services.NewVolumeService(db, svcs.Docker, svcs.Event)
	svcs.Network = services.NewNetworkService(db, svcs.Docker, svcs.Event)
	svcs.Template = services.NewTemplateService(ctx, db, httpClient, svcs.Settings)
	svcs.Auth = services.NewAuthService(svcs.User, svcs.Settings, svcs.Event, cfg.JWTSecret, cfg)
	svcs.Oidc = services.NewOidcService(svcs.Auth, cfg, httpClient)
	svcs.Updater = services.NewUpdaterService(db, svcs.Settings, svcs.Docker, svcs.Project, svcs.ImageUpdate, svcs.ContainerRegistry, svcs.Event, svcs.Image)
	svcs.System = services.NewSystemService(db, svcs.Docker, svcs.Container, svcs.Image, svcs.Volume, svcs.Network, svcs.Settings)
	svcs.Version = services.NewVersionService(httpClient, cfg.UpdateCheckDisabled, config.Version, config.Revision)
	svcs.SSH = services.NewSSHService()

	return svcs, dockerClient, nil
}
