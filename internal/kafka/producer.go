package kafka

import (
	"encoding/json"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)


type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaProducer(broker, topic string) (*KafkaProducer, error) {
	compression := os.Getenv("KAFKA_COMPRESSION")
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"compression.type":  compression,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		producer: p,
		topic:    topic,
	}, nil
}

func (kp *KafkaProducer) Publish(message interface{}, eventType string) error {
	// Marshal pesan ke JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Failed to marshal message: %v", err)
		return err
	}

	// Buat pesan Kafka dengan raw JSON (byte slice)
	return kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kp.topic,
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
