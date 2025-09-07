//go:build playwright

package bootstrap

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/ofkm/arcane-backend/internal/api"
	"github.com/ofkm/arcane-backend/internal/services"
)

func init() {
	registerPlaywrightRoutes = []func(apiGroup *gin.RouterGroup, services *Services){
		func(apiGroup *gin.RouterGroup, svc *Services) {
			playwrightService := services.NewPlaywrightService(svc.Settings)
			if playwrightService == nil {
				slog.Warn("Playwright service not available, skipping playwright routes")
				return
			}

			api.SetupPlaywrightRoutes(apiGroup, playwrightService)
			slog.Info("Playwright routes registered for E2E testing")
		},
	}
}
