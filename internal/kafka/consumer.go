package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"go-crud/internal/entity"
	"go-crud/internal/repository"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConsumer struct {
	consumer     *kafka.Consumer
	userRepo     repository.UserRepository
	repoRepo     repository.RepositoryRepository
}

func NewKafkaConsumer(broker, groupID, topic string, userRepo repository.UserRepository, repoRepo repository.RepositoryRepository) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: c,
		userRepo: userRepo,
		repoRepo: repoRepo,
	}, nil
}

func (kc *KafkaConsumer) Start(ctx context.Context) {
	log.Println("🚀 Kafka consumer started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Kafka consumer stopped")
			kc.consumer.Close()
			return
		default:
			msg, err := kc.consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("⚠️ Error reading message: %v\n", err)
				continue
			}

			var event map[string]interface{}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("⚠️ Failed to unmarshal message: %v\n", err)
				continue
			}

			kc.processEvent(ctx, event)
		}
	}
}

func (kc *KafkaConsumer) processEvent(ctx context.Context, event map[string]interface{}) {
	eventType := fmt.Sprintf("%v", event["event"])
	log.Printf("📥 Processing event: %s\n", eventType)

	switch eventType {
	// User events
	case "user.created":
		user := entity.User{
			Name:  fmt.Sprintf("%v", event["name"]),
			Email: fmt.Sprintf("%v", event["email"]),
		}
		if err := kc.userRepo.CreateUser(ctx, &user); err != nil {
			log.Printf("❌ Failed to create user from event: %v\n", err)
		}

	case "user.updated":
		id := toInt(event["id"])
		user, err := kc.userRepo.GetUserByID(ctx, id)
		if err != nil {
			log.Printf("❌ User not found for update: %v\n", err)
			return
		}
		user.Name = fmt.Sprintf("%v", event["name"])
		user.Email = fmt.Sprintf("%v", event["email"])
		kc.userRepo.UpdateUser(ctx, user)

	case "user.deleted":
		id := toInt(event["id"])
		kc.userRepo.DeleteUser(ctx, id)

	// Repository events
	case "repository.created":
		repo := entity.Repository{
			Name:      fmt.Sprintf("%v", event["name"]),
			URL:       fmt.Sprintf("%v", event["url"]),
			AIEnabled: toBool(event["ai_enabled"]),
			UserID:    toInt(event["user_id"]),
		}
		if err := kc.repoRepo.CreateRepository(ctx, &repo); err != nil {
			log.Printf("❌ Failed to create repository from event: %v\n", err)
		}

	case "repository.updated":
		id := toInt(event["id"])
		repo, err := kc.repoRepo.GetRepositoryByID(ctx, id)
		if err != nil {
			log.Printf("❌ Repository not found for update: %v\n", err)
			return
		}
		repo.Name = fmt.Sprintf("%v", event["name"])
		repo.URL = fmt.Sprintf("%v", event["url"])
		repo.AIEnabled = toBool(event["ai_enabled"])
		kc.repoRepo.Update(ctx, repo)

	case "repository.deleted":
		id := toInt(event["id"])
		kc.repoRepo.Delete(ctx, id)

	default:
		log.Printf("⚠️ Unknown event type: %s\n", eventType)
	}
}

// Helper for type conversion
func toInt(val interface{}) int {
	if f, ok := val.(float64); ok {
		return int(f)
	}
	return 0
}

func toBool(val interface{}) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}
