package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"go-crud/internal/entity"
	"go-crud/internal/usecase"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConsumer struct {
	consumer        *kafka.Consumer
	userUsecase     usecase.IUserUsecase
	repoUsecase     usecase.IRepositoryUsecase
}

func NewKafkaConsumer(
	broker, groupID, topic string,
	userUC usecase.IUserUsecase,
	repoUC usecase.IRepositoryUsecase,
) (*KafkaConsumer, error){
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  broker,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest", // atau "latest" tergantung kebutuhan
		"enable.auto.commit": true,       // ✅ AUTO COMMIT DIHIDUPKAN
		"auto.commit.interval.ms": 5000,  // (opsional) commit tiap 5 detik
	})
	
	
	if err != nil {
		log.Printf("❌ Error creating consumer: %v\n", err)
		return nil, err
	}

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		log.Printf("❌ Error subscribing to topic: %v\n", err)
		return nil, err
	}

	log.Printf("✅ Kafka consumer subscribed to topic: %s\n", topic)



	return &KafkaConsumer{
		consumer:    c,
		userUsecase: userUC,
		repoUsecase: repoUC,
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

			log.Printf("📨 Received message from topic: %s\n", *msg.TopicPartition.Topic)
			log.Printf("📥 Message value: %s\n", string(msg.Value))


			var event map[string]interface{}
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("⚠️ Failed to unmarshal message: %v\n", err)
				continue
			}

			topic := *msg.TopicPartition.Topic
			log.Printf("📋 Routing event by topic: %s\n", topic)
			kc.routeEventByTopic(ctx, topic, event)
		}
	}
}

func (kc *KafkaConsumer) routeEventByTopic(ctx context.Context, topic string, event map[string]interface{}) {
	log.Printf("📥 Processing event from topic: %s\n", topic)
	log.Printf("📥 Processing event from topic: %s\n", topic)
	log.Printf("🧾 Event payload received: %+v\n", event) 

	switch topic {
	case "user-events":
		kc.processUserEvent(ctx, event)

	case "repository-events":
		kc.processRepositoryEvent(ctx, event)

	default:
		log.Printf("⚠️ Unknown topic: %s\n", topic)
	}
}


func isUserEvent(eventType string) bool {
	return eventType == "user.created" || eventType == "user.updated" || eventType == "user.deleted"
}

func isRepoEvent(eventType string) bool {
	return eventType == "repository.created" || eventType == "repository.updated" || eventType == "repository.deleted"
}

func (kc *KafkaConsumer) processUserEvent(ctx context.Context, event map[string]interface{}) {
	eventType := fmt.Sprintf("%v", event["event"])
	log.Printf("🔍 Handling user event type: %s | Data: %+v\n", eventType, event)

	switch eventType {
	case "user.created":
		user := &entity.User{
			Name:  fmt.Sprintf("%v", event["name"]),
			Email: fmt.Sprintf("%v", event["email"]),
		}
		err := kc.userUsecase.CreateUser(ctx, user)
		if err != nil {
			log.Printf("❌ Failed to create user from event: %v\n", err)
		}		

	case "user.updated":
		id := toInt(event["id"])
		input := usecase.UserInput{
			Name:  fmt.Sprintf("%v", event["name"]),
			Email: fmt.Sprintf("%v", event["email"]),
		}
		_, err := kc.userUsecase.UpdateUser(ctx, id, input)
		if err != nil {
			log.Printf("❌ Failed to update user from event: %v\n", err)
		}

	case "user.deleted":
		id := toInt(event["id"])
		err := kc.userUsecase.DeleteUser(ctx, id)
		if err != nil {
			log.Printf("❌ Failed to delete user from event: %v\n", err)
		}

	default:
		log.Printf("⚠️ Unknown user event: %s\n", eventType)
	}
}


func (kc *KafkaConsumer) processRepositoryEvent(ctx context.Context, event map[string]interface{}) {
	eventType := fmt.Sprintf("%v", event["event"])

	switch eventType {
	case "repository.created":
		repoInput := entity.Repository{
			Name:      fmt.Sprintf("%v", event["name"]),
			URL:       fmt.Sprintf("%v", event["url"]),
			AIEnabled: toBool(event["ai_enabled"]),
			UserID:    toInt(event["user_id"]),
		}
		err := kc.repoUsecase.CreateRepository(ctx, &repoInput)
		if err != nil {
			log.Printf("❌ Failed to create repository from event: %v\n", err)
		}

	case "repository.updated":
		id := toInt(event["id"])
		repoInput := usecase.RepositoryInput{
			Name:      fmt.Sprintf("%v", event["name"]),
			URL:       fmt.Sprintf("%v", event["url"]),
			AIEnabled: toBool(event["ai_enabled"]),
		}
		_, err := kc.repoUsecase.UpdateRepository(ctx, id, repoInput)
		if err != nil {
			log.Printf("❌ Failed to update repository from event: %v\n", err)
		}

	case "repository.deleted":
		id := toInt(event["id"])
		err := kc.repoUsecase.DeleteRepository(ctx, id)
		if err != nil {
			log.Printf("❌ Failed to delete repository from event: %v\n", err)
		}

	default:
		log.Printf("⚠️ Unknown repository event: %s\n", eventType)
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



// func (kc *KafkaConsumer) processEvent(ctx context.Context, event map[string]interface{}) {
// 	eventType := fmt.Sprintf("%v", event["event"])
// 	log.Printf("📥 Processing event: %s\n", eventType)

// 	switch eventType {
// 	// User events
// 	case "user.created":
// 		user := entity.User{
// 			Name:  fmt.Sprintf("%v", event["name"]),
// 			Email: fmt.Sprintf("%v", event["email"]),
// 		}
// 		if err := kc.userRepo.CreateUser(ctx, &user); err != nil {
// 			log.Printf("❌ Failed to create user from event: %v\n", err)
// 		}

// 	case "user.updated":
// 		id := toInt(event["id"])
// 		user, err := kc.userRepo.GetUserByID(ctx, id)
// 		if err != nil {
// 			log.Printf("❌ User not found for update: %v\n", err)
// 			return
// 		}
// 		user.Name = fmt.Sprintf("%v", event["name"])
// 		user.Email = fmt.Sprintf("%v", event["email"])
// 		kc.userRepo.UpdateUser(ctx, user)

// 	case "user.deleted":
// 		id := toInt(event["id"])
// 		kc.userRepo.DeleteUser(ctx, id)

// 	// Repository events
// 	case "repository.created":
// 		repo := entity.Repository{
// 			Name:      fmt.Sprintf("%v", event["name"]),
// 			URL:       fmt.Sprintf("%v", event["url"]),
// 			AIEnabled: toBool(event["ai_enabled"]),
// 			UserID:    toInt(event["user_id"]),
// 		}
// 		if err := kc.repoRepo.CreateRepository(ctx, &repo); err != nil {
// 			log.Printf("❌ Failed to create repository from event: %v\n", err)
// 		}

// 	case "repository.updated":
// 		id := toInt(event["id"])
// 		repo, err := kc.repoRepo.GetRepositoryByID(ctx, id)
// 		if err != nil {
// 			log.Printf("❌ Repository not found for update: %v\n", err)
// 			return
// 		}
// 		repo.Name = fmt.Sprintf("%v", event["name"])
// 		repo.URL = fmt.Sprintf("%v", event["url"])
// 		repo.AIEnabled = toBool(event["ai_enabled"])
// 		kc.repoRepo.Update(ctx, repo)

// 	case "repository.deleted":
// 		id := toInt(event["id"])
// 		kc.repoRepo.Delete(ctx, id)

// 	default:
// 		log.Printf("⚠️ Unknown event type: %s\n", eventType)
// 	}
// }