package models

import (
	"time"
)

type ContentDocument struct {
	ID        string            `json:"id"`
	UserID    int               `json:"user_id,omitempty"`
	Title     string            `json:"title" binding:"required"`
	Body      string            `json:"body" binding:"required"`
	Tags      []string          `json:"tags"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	UpdatedAt time.Time         `json:"updated_at"`
}