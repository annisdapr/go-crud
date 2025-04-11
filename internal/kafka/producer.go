package kafka
import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaProducer(broker, topic string) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		producer: p,
		topic:    topic,
	}, nil
}

func (kp *KafkaProducer) Publish(message string) error {
	return kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &kp.topic,
			Partition: int32(kafka.PartitionAny), // ✅ Cast ke int32
		},
		Value: []byte(message),
	}, nil)
}

func (kp *KafkaProducer) Close() {
	kp.producer.Close()
}
