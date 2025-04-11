package repository

import (
	"context"
	"go-crud/internal/entity"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuditLogMongoRepository interface {
	InsertLog(ctx context.Context, log *entity.AuditLog) error
	GetLogs(ctx context.Context, userID int) ([]entity.AuditLog, error)
}

type auditLogRepo struct {
	collection *mongo.Collection
}

func NewAuditLogMongoRepository(db *mongo.Database) AuditLogMongoRepository {
	return &auditLogRepo{
		collection: db.Collection("audit_logs"),
	}
}

func (r *auditLogRepo) InsertLog(ctx context.Context, log *entity.AuditLog) error {
	log.Timestamp = time.Now()
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

func (r *auditLogRepo) GetLogs(ctx context.Context, userID int) ([]entity.AuditLog, error) {
	var results []entity.AuditLog
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var log entity.AuditLog
		if err := cursor.Decode(&log); err != nil {
			return nil, err
		}
		results = append(results, log)
	}
	return results, nil
}
