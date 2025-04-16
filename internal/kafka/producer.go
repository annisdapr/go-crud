package kafka

import (
	"encoding/json"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer(broker string) (*KafkaProducer, error) {
	compression := os.Getenv("KAFKA_COMPRESSION")
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"compression.type":  compression,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer: p}, nil
}

func (kp *KafkaProducer) Publish(topic string, message interface{}, eventType string) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Failed to marshal message: %v", err)
		return err
	}

	return kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: int32(kafka.PartitionAny),
		},
		Value: messageBytes,
		Headers: []kafka.Header{
			{Key: "eventType", Value: []byte(eventType)},
		},
	}, nil)
}

func (kp *KafkaProducer) Close() {
	kp.producer.Close()
}
