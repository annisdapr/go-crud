package entity

import (
    "time"
)

type Repository struct {
    ID        int       `json:"id,omitempty"`
    UserID    int       `json:"user_id"`    // Relasi ke User
    Name      string    `json:"name" validate:"required"`
    URL       string    `json:"url" validate:"required,url"`
    AIEnabled bool      `json:"ai_enabled"` 
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
