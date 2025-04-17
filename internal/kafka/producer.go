package kafka

import (
	"encoding/json"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaProducer struct {
	Producer *kafka.Producer
}

func NewKafkaProducer(broker string) (*KafkaProducer, error) {
	compression := os.Getenv("KAFKA_COMPRESSION")
	brokerKafka := os.Getenv("KAFKA_BROKER")
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokerKafka,
		"compression.type":  compression,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{Producer: p}, nil
}

func (kp *KafkaProducer) Publish(topic string, message interface{}, eventType string) error {
	payload := make(map[string]interface{})

	// Merge isi message ke payload
	if msgMap, ok := message.(map[string]interface{}); ok {
		for k, v := range msgMap {
			payload[k] = v
		}
	} else {
		messageBytes, _ := json.Marshal(message)
		json.Unmarshal(messageBytes, &payload)
	}

	// Set eventType setelah merge (jadi tidak bisa ditimpa)
	payload["event"] = eventType

	finalBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Failed to marshal final payload: %v", err)
		return err
	}

	log.Printf("🧪 Final payload before sending to Kafka: %s\n", string(finalBytes))

	return kp.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: int32(kafka.PartitionAny),
		},
		Value: finalBytes,
		Headers: []kafka.Header{
			{Key: "eventType", Value: []byte(eventType)},
		},
	}, nil)
}

func (kp *KafkaProducer) Close() {
	kp.Producer.Close()
}
// func (kp *KafkaProducer) Publish(topic string, message interface{}, eventType string) error {
// 	// Pastikan eventType ikut dimasukkan ke dalam message
// 	payload := map[string]interface{}{
// 		"event": eventType,
// 	}

// 	// Ambil isi message dan merge ke payload
// 	messageBytes, _ := json.Marshal(message)
// 	json.Unmarshal(messageBytes, &payload)

// 	finalBytes, err := json.Marshal(payload)
// 	if err != nil {
// 		log.Printf("❌ Failed to marshal final payload: %v", err)
// 		return err
// 	}

// 	return kp.Producer.Produce(&kafka.Message{
// 		TopicPartition: kafka.TopicPartition{
// 			Topic:     &topic,
// 			Partition: int32(kafka.PartitionAny),
// 		},
// 		Value: finalBytes,
// 		Headers: []kafka.Header{
// 			{Key: "eventType", Value: []byte(eventType)},
// 		},
// 	}, nil)
// }

// func (kp *KafkaProducer) Close() {
// 	kp.Producer.Close()
// }