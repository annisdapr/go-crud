package port

import (
	"context"
	"go-crud/internal/entity"
)

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, user *entity.User) error
}
