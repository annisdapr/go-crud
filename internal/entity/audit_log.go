package entity

import "time"

type AuditLog struct {
	UserID    int       `bson:"user_id"`
	UserName  string    `bson:"user_name"` 
	Action    string    `bson:"action"`
	Timestamp time.Time `bson:"timestamp"`
}
