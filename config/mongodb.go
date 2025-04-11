package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB(uri string, dbName string) (*mongo.Database, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	clientOpts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	// Test koneksi
	if err := client.Ping(ctx, nil); err != nil {
		cancel()
		return nil, nil, err
	}

	log.Println("Connected to MongoDB")

	cleanup := func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println("Failed to disconnect MongoDB:", err)
		} else {
			log.Println("MongoDB disconnected")
		}
		cancel()
	}

	return client.Database(dbName), cleanup, nil
}
