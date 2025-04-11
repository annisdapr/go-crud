package entity

import (
    "time"
)

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name" validate:"required,min=3"`
    Email     string    `json:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}