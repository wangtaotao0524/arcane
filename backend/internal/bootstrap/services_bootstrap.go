package bootstrap

import (
	"log"

	"github.com/ofkm/arcane-backend/internal/api"
	"github.com/ofkm/arcane-backend/internal/config"
	"github.com/ofkm/arcane-backend/internal/database"
	"github.com/ofkm/arcane-backend/internal/services"
)

func initializeServices(db *database.DB, cfg *config.Config) (*api.Services, *services.DockerClientService, error) {
	log.Println("Initializing services...")
	eventService := services.NewEventService(db)
	converterService := services.NewConverterService()
	settingsService := services.NewSettingsService(db, cfg)
	dockerClientService := services.NewDockerClientService(db)
	userService := services.NewUserService(db)
	containerRegistry := services.NewContainerRegistryService(db)
	imageUpdate := services.NewImageUpdateService(db, settingsService, containerRegistry, dockerClientService, eventService)
	imageService := services.NewImageService(db, dockerClientService, containerRegistry, imageUpdate, eventService)
	stackService := services.NewStackService(db, settingsService, eventService, imageService)
	environmentService := services.NewEnvironmentService(db)
	containerService := services.NewContainerService(db, eventService, dockerClientService)
	volumeService := services.NewVolumeService(db, dockerClientService, eventService)
	networkService := services.NewNetworkService(db, dockerClientService, eventService)
	templateService := services.NewTemplateService(db)
	authService := services.NewAuthService(userService, settingsService, eventService, cfg.JWTSecret, cfg)
	oidcService := services.NewOidcService(authService, cfg)
	updaterService := services.NewUpdaterService(db, settingsService, dockerClientService, stackService, imageUpdate, containerRegistry, eventService, imageService)
	systemService := services.NewSystemService(db, dockerClientService, containerService, imageService, volumeService, networkService, settingsService)

	appServices := &api.Services{
		User:              userService,
		Stack:             stackService,
		Environment:       environmentService,
		Settings:          settingsService,
		Container:         containerService,
		Image:             imageService,
		Volume:            volumeService,
		Network:           networkService,
		Auth:              authService,
		Oidc:              oidcService,
		Docker:            dockerClientService,
		Converter:         converterService,
		Template:          templateService,
		ContainerRegistry: containerRegistry,
		System:            systemService,
		Updater:           updaterService,
		ImageUpdate:       imageUpdate,
		Event:             eventService,
	}

	return appServices, dockerClientService, nil
}
