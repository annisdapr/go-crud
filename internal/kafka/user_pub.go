// package kafka

// import (
// 	"context"
// 	"encoding/json"
// 	"go-crud/internal/entity"
// 	"go-crud/internal/usecase/port"

// 	confluentKafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
// )

// type KafkaUserPublisher struct {
// 	Producer *confluentKafka.Producer
// }

// 


// 	return k.Producer.Produce(msg, nil)
// }

// internal/kafka/user_event_publisher.go
package kafka

import (
	"context"
	"encoding/json"
	"go-crud/internal/entity"
	"go-crud/internal/usecase/port"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaUserPublisher struct {
	Producer *confluentKafka.Producer
}

func NewKafkaUserPublisher(producer  *confluentKafka.Producer) port.EventPublisher {
	return &KafkaUserPublisher{Producer: producer}
}

func (k *KafkaUserPublisher) PublishUserCreated(ctx context.Context, user *entity.User) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	topic := "user-events"
	msg := &confluentKafka.Message{
		TopicPartition: confluentKafka.TopicPartition{
			Topic:     &topic,
			Partition: int32(confluentKafka.PartitionAny),
		},
		Value: payload,
		Headers: []confluentKafka.Header{
			{
				Key:   "eventType",
				Value: []byte("user.created"),
			},
		},
	}
	return k.Producer.Produce(msg, nil)
}

// func (k *KafkaUserPublisher) PublishUserCreated(ctx context.Context, user *entity.User) error {
// 	return k.Producer.Publish("user-events", user, "user.created")
// }
