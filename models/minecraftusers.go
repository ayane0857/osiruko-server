package models

import (
	"time"

	"github.com/google/uuid"
)

type Users struct {
	ID               uint           `gorm:"primary_key"`
	MinecraftUUID    uuid.UUID      `gorm:"type:uuid;uniqueIndex"`
	Name			 string         `gorm:"size:255"`
	Email			 string         `gorm:"size:255;uniqueIndex"`
	CreatedAt        time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
}