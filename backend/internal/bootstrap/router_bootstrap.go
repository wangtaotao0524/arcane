package bootstrap

import (
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"

	"github.com/ofkm/arcane-backend/frontend"
	"github.com/ofkm/arcane-backend/internal/api"
	"github.com/ofkm/arcane-backend/internal/config"
	"github.com/ofkm/arcane-backend/internal/middleware"
)

var registerPlaywrightRoutes []func(apiGroup *gin.RouterGroup, services *api.Services)

func setupRouter(cfg *config.Config, appServices *api.Services) *gin.Engine {

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())

	loggerSkipPatterns := []string{
		"GET /api/containers/*/stats/stream",
		"GET /api/containers/*/logs/stream",
		"GET /_app",
		"GET /img",
		"GET /fonts",
		"GET /api/system/stats",
		"GET /health",
		"HEAD /health",
	}
	router.Use(sloggin.NewWithConfig(slog.Default(), sloggin.Config{
		Filters: []sloggin.Filter{
			func(c *gin.Context) bool {
				mp := c.Request.Method + " " + c.Request.URL.Path
				for _, pat := range loggerSkipPatterns {
					if pat == mp {
						return false
					}
					if ok, _ := path.Match(pat, mp); ok {
						return false
					}
					if strings.HasSuffix(pat, "/") && strings.HasPrefix(mp, pat) {
						return false
					}
				}
				return true
			},
		},
	}))

	router.Use(middleware.SetupCORS(cfg))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	api.SetupRoutes(router, appServices, cfg)

	// Register playwright routes if available
	if registerPlaywrightRoutes != nil {
		apiGroup := router.Group("/api")
		for _, registerFunc := range registerPlaywrightRoutes {
			registerFunc(apiGroup, appServices)
		}
	}

	if err := frontend.RegisterFrontend(router); err != nil {
		_, _ = gin.DefaultErrorWriter.Write([]byte("Failed to register frontend: " + err.Error() + "\n"))
	}

	return router
}
