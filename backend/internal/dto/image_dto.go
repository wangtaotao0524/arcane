package dto

import (
	"time"

	"github.com/ofkm/arcane-backend/internal/models"
)

type ImagePullDto struct {
	ImageName string `json:"imageName" binding:"required"`
}

type ImageUpdateInfoDto struct {
	HasUpdate      bool      `json:"hasUpdate"`
	UpdateType     string    `json:"updateType"`
	CurrentVersion string    `json:"currentVersion"`
	LatestVersion  string    `json:"latestVersion"`
	CurrentDigest  string    `json:"currentDigest"`
	LatestDigest   string    `json:"latestDigest"`
	CheckTime      time.Time `json:"checkTime"`
	ResponseTimeMs int       `json:"responseTimeMs"`
	Error          string    `json:"error"`
}

type ImageSummaryDto struct {
	ID          string                 `json:"id" sortable:"true"`
	RepoTags    []string               `json:"repoTags"`
	RepoDigests []string               `json:"repoDigests"`
	Created     int64                  `json:"created" sortable:"true"`
	Size        int64                  `json:"size" sortable:"true"`
	VirtualSize int64                  `json:"virtualSize"`
	Labels      map[string]interface{} `json:"labels"`
	InUse       bool                   `json:"inUse" sortable:"true"`
	Repo        string                 `json:"repo" sortable:"true"`
	Tag         string                 `json:"tag" sortable:"true"`
	UpdateInfo  *ImageUpdateInfoDto    `json:"updateInfo,omitempty"`
}

func NewImageSummaryDto(m *models.Image) ImageSummaryDto {
	var labels map[string]interface{}
	if m.Labels != nil {
		labels = map[string]interface{}(m.Labels)
	}

	var update *ImageUpdateInfoDto
	if m.UpdateRecord != nil {
		sp := func(p *string) string {
			if p == nil {
				return ""
			}
			return *p
		}

		update = &ImageUpdateInfoDto{
			HasUpdate:      m.UpdateRecord.HasUpdate,
			UpdateType:     m.UpdateRecord.UpdateType,
			CurrentVersion: m.UpdateRecord.CurrentVersion,
			LatestVersion:  sp(m.UpdateRecord.LatestVersion),
			CurrentDigest:  sp(m.UpdateRecord.CurrentDigest),
			LatestDigest:   sp(m.UpdateRecord.LatestDigest),
			CheckTime:      m.UpdateRecord.CheckTime,
			ResponseTimeMs: m.UpdateRecord.ResponseTimeMs,
			Error:          sp(m.UpdateRecord.LastError),
		}
	}

	return ImageSummaryDto{
		ID:          m.ID,
		RepoTags:    []string(m.RepoTags),
		RepoDigests: []string(m.RepoDigests),
		Created:     m.Created.Unix(),
		Size:        m.Size,
		VirtualSize: m.VirtualSize,
		Labels:      labels,
		InUse:       m.InUse,
		Repo:        m.Repo,
		Tag:         m.Tag,
		UpdateInfo:  update,
	}
}
