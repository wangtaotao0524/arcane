package models

import "time"

type Volume struct {
	Name       string    `json:"name" gorm:"primaryKey;type:text"`
	Driver     string    `json:"driver" gorm:"not null"`
	Mountpoint string    `json:"mountpoint" gorm:"not null"`
	Labels     JSON      `json:"labels,omitempty" gorm:"type:text"`
	Scope      string    `json:"scope" gorm:"not null"`
	Options    JSON      `json:"options,omitempty" gorm:"type:text"`
	CreatedAt  time.Time `json:"createdAt" gorm:"not null"`

	InUse bool `json:"inUse" gorm:"column:in_use;default:false"`

	BaseModel
}

func (Volume) TableName() string {
	return "volumes_table"
}
