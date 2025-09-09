package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/docker/docker/api/types/image"
	"github.com/gin-gonic/gin"
	"github.com/ofkm/arcane-backend/internal/dto"
	"github.com/ofkm/arcane-backend/internal/middleware"
	"github.com/ofkm/arcane-backend/internal/services"
	"github.com/ofkm/arcane-backend/internal/utils"
)

type ImageHandler struct {
	imageService       *services.ImageService
	imageUpdateService *services.ImageUpdateService
}

func NewImageHandler(group *gin.RouterGroup, imageService *services.ImageService, imageUpdateService *services.ImageUpdateService, authMiddleware *middleware.AuthMiddleware) {
	handler := &ImageHandler{imageService: imageService, imageUpdateService: imageUpdateService}

	apiGroup := group.Group("/images")
	apiGroup.Use(authMiddleware.WithAdminNotRequired().Add())
	{
		apiGroup.GET("", handler.List)
		apiGroup.GET("/:id", handler.GetByID)
		apiGroup.DELETE("/:id", handler.Remove)
		apiGroup.POST("/pull", handler.Pull)
		apiGroup.POST("/prune", handler.Prune)
		apiGroup.GET("/:id/history", handler.GetHistory)
		apiGroup.GET("/total-size", handler.GetTotalSize)
	}
}

func (h *ImageHandler) List(c *gin.Context) {
	var req utils.SortedPaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: "Invalid pagination or sort parameters: " + err.Error()},
		})
		return
	}

	if req.Pagination.Page == 0 {
		req.Pagination.Page = 1
	}
	if req.Pagination.Limit == 0 {
		req.Pagination.Limit = 20
	}

	images, pagination, err := h.imageService.ListImagesWithUpdatesPaginated(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: "Failed to list images: " + err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       images,
		"pagination": pagination,
	})
}

func (h *ImageHandler) GetByID(c *gin.Context) {
	id := c.Param("imageId")

	img, err := h.imageService.GetImageByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "data": dto.MessageDto{Message: err.Error()}})
		return
	}

	out := dto.NewImageDetailSummaryDto(img)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    out,
	})
}

func (h *ImageHandler) Remove(c *gin.Context) {
	id := c.Param("imageId")
	force := c.Query("force") == "true"

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "data": dto.MessageDto{Message: "User not authenticated"}})
		return
	}
	if err := h.imageService.RemoveImage(c.Request.Context(), id, force, *currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.MessageDto{Message: "Image removed successfully"},
	})
}

func (h *ImageHandler) Pull(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.ImagePullDto

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: "Invalid request body: " + err.Error()},
		})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/x-json-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists || currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "data": dto.MessageDto{Message: "User not authenticated"}})
		return
	}

	if err := h.imageService.PullImage(ctx, req.ImageName, c.Writer, *currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: fmt.Sprintf("Failed to pull image '%s': %s", req.ImageName, err.Error())},
		})
		return
	}

	slog.InfoContext(ctx, "Image pull stream completed and database sync attempted",
		slog.String("imageName", req.ImageName))
}

func (h *ImageHandler) Prune(c *gin.Context) {
	dangling := c.Query("dangling") == "true"

	report, err := h.imageService.PruneImages(c.Request.Context(), dangling)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: err.Error()},
		})
		return
	}

	out := dto.NewImagePruneReportDto(*report)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    out,
	})
}

func (h *ImageHandler) GetHistory(c *gin.Context) {
	id := c.Param("id")

	history, err := h.imageService.GetImageHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: err.Error()},
		})
		return
	}

	out, mapErr := dto.MapSlice[image.HistoryResponseItem, dto.ImageHistoryItemDto](history)
	if mapErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "data": dto.MessageDto{Message: mapErr.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    out,
	})
}

func (h *ImageHandler) GetTotalSize(c *gin.Context) {
	ctx := c.Request.Context()

	total, err := h.imageService.GetTotalImageSize(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data":    dto.MessageDto{Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dto.TotalImageSizeDto{TotalSize: total},
	})
}
