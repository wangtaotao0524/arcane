package models

import "time"

type StackStatus string

const (
	StackStatusRunning          StackStatus = "running"
	StackStatusStopped          StackStatus = "stopped"
	StackStatusPartiallyRunning StackStatus = "partially running"
	StackStatusUnknown          StackStatus = "unknown"
	StackStatusDeploying        StackStatus = "deploying"
	StackStatusStopping         StackStatus = "stopping"
	StackStatusRestarting       StackStatus = "restarting"
)

type Stack struct {
	Name         string      `json:"name" gorm:"not null" sortable:"true"`
	DirName      *string     `json:"dir_name" gorm:"unique"`
	Path         string      `json:"path" gorm:"not null"`
	Status       StackStatus `json:"status" sortable:"true"`
	ServiceCount int         `json:"service_count" sortable:"true"`
	RunningCount int         `json:"running_count" sortable:"true"`
	AutoUpdate   bool        `json:"auto_update" sortable:"true"`
	IsExternal   bool        `json:"is_external"`
	IsLegacy     bool        `json:"is_legacy"`
	IsRemote     bool        `json:"is_remote"`

	BaseModel
}

func (Stack) TableName() string {
	return "stacks"
}

type ProjectCache struct {
	BaseModel
	StackID      string      `json:"stack_id" gorm:"uniqueIndex;not null"`
	Name         string      `json:"name" gorm:"not null"`
	Status       StackStatus `json:"status" gorm:"not null"`
	ServiceCount int         `json:"service_count"`
	RunningCount int         `json:"running_count"`
	AutoUpdate   bool        `json:"auto_update"`
	LastModified time.Time   `json:"last_modified"`
	ComposeHash  string      `json:"compose_hash"`
	CachedAt     time.Time   `json:"cached_at"`
}

func (ProjectCache) TableName() string {
	return "project_cache"
}
