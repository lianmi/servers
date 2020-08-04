package kafka

import (
	// "encoding/hex"
	"encoding/json"
	"strings"

	"os"
	"os/signal"
	"syscall"

	// "time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/lianmi/servers/internal/app/channel"
	"github.com/lianmi/servers/internal/pkg/models"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type KafkaOptions struct {
	Broker string
	Group  string
	Topics string //以逗号隔开: Auth.Frontend Users.Frontend ... etc.
}

type KafkaClient struct {
	o                *KafkaOptions
	app              string
	topics           []string
	kafkamqttChannel *channel.KafkaMqttChannel
	consumer         *kafka.Consumer
	producer         *kafka.Producer
	logger           *zap.Logger
	redisPool        *redis.Pool
}

func NewKafkaOptions(v *viper.Viper) (*KafkaOptions, error) {
	var (
		err error
		o   = new(KafkaOptions)
	)

	if err = v.UnmarshalKey("kafka", o); err != nil {
		return nil, err
	}

	return o, err
}

func NewKafkaClient(o *KafkaOptions, redisPool *redis.Pool, channel *channel.KafkaMqttChannel, logger *zap.Logger) *KafkaClient {
	topicArray := strings.Split(o.Topics, ",")
	topics := make([]string, 0)
	for _, topic := range topicArray {
		topics = append(topics, topic)
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               o.Broker,
		"group.id":                        o.Group,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		// Enable generation of PartitionEOF when the
		// end of a partition is reached.
		"enable.partition.eof": true,
		"auto.offset.reset":    "earliest"})

	if err != nil {
		logger.Error("Failed to create consumer, retry... ", zap.Error(err))
		return nil
	}
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": o.Broker})

	if err != nil {
		logger.Error("Failed to create producer: ", zap.Error(err))
		return nil
	}
	return &KafkaClient{
		o:                o,
		topics:           topics,
		kafkamqttChannel: channel,
		consumer:         c,
		producer:         p,
		logger:           logger.With(zap.String("type", "kafka.Client")),
		redisPool:        redisPool,
	}
}

func (kc *KafkaClient) Application(name string) {
	kc.app = name
}

//启动Kafka实例
func (kc *KafkaClient) Start() error {
	kc.logger.Info("==> Subscribe Topics ", zap.String("Topics", kc.o.Topics))

	err := kc.consumer.SubscribeTopics(kc.topics, nil)
	if err != nil {
		kc.logger.Error("Failed to SubscribeTopics: ", zap.Error(err))
		return err
	}

	//尝试读取redis
	redisConn := kc.redisPool.Get()
	// defer redisConn.Close()
	// vkey := fmt.Sprintf("verificationCode:%s", email)

	if bar, err := redis.String(redisConn.Do("GET", "bar")); err == nil {
		kc.logger.Info("redisConn GET ", zap.String("bar", bar))
	}

	//Go程
	go kc.ProcessProduceChan()

	go func() {
		run := true

		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		for run == true {
			select {
			case sig := <-sigchan:
				kc.logger.Info("Caught signal terminating")
				_ = sig
				run = false

			case ev := <-kc.consumer.Events(): //收到后端服务端的kafka消息
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					kc.logger.Info("AssignedPartitions")
					kc.consumer.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					kc.logger.Info("RevokedPartitions")
					kc.consumer.Unassign()
				case *kafka.Message:
					topic := *e.TopicPartition.Topic
					kc.logger.Info("kafka.Message Reviced from authService", zap.String("Topic:", topic))

					backendMessage := new(models.Message)

					if err := json.Unmarshal(e.Value, backendMessage); err == nil {

						businessType := backendMessage.GetBusinessType()
						businessSubType := backendMessage.GetBusinessSubType()
						taskId := backendMessage.GetTaskId()
						businessTypeName := backendMessage.GetBusinessTypeName()
						code := backendMessage.GetCode()

						//根据目标target,  组装数据包， 向mqtt的channel写入
						kc.logger.Info("Message on backend service",
							zap.String("Topic:", topic),
							// zap.String("Account:", backendMessage.GetAccount()),
							zap.Uint32("taskId:", taskId),
							zap.String("BusinessTypeName:", businessTypeName), //业务
							zap.Uint32("businessType:", businessType),         // 业务类型
							zap.Uint32("businessSubType:", businessSubType),   // 业务子类型
							zap.Int32("code:", code),   // 状态码
							zap.String("Source:", backendMessage.GetSource()), //发送者
							zap.String("Target:", backendMessage.GetTarget()), //接收者
						)

						//向MTChan通道写入数据, 实现向客户端发送数据
						kc.kafkamqttChannel.MTChan <- backendMessage
					} else {
						continue
					}

				case kafka.PartitionEOF:
					kc.logger.Info("kafka.PartitionEOF")

				case kafka.Error:
					// Errors should generally be considered as informational, the client will try to automatically recover
					kc.logger.Info("kafka.Error")
				}
			}
		}

		kc.logger.Info("Closing consumer")
		kc.consumer.Close()
	}()

	return nil
}

// 处理生产者数据，这些数据来源于mqtt的订阅消费，向后端场景服务程序发布
func (kc *KafkaClient) ProcessProduceChan() {
	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			kc.logger.Info("Caught signal terminating")
			_ = sig
			run = false
		case msg := <-kc.kafkamqttChannel.KafkaChan: //从KafkaChan通道里读取数据
			//Target字段存储的是业务模块名称: "auth.Backend", "users.Backend" ...
			topic := msg.GetTarget()
			kc.Produce(topic, msg)
		}
	}
}

//Produce
func (kc *KafkaClient) Produce(topic string, msg *models.Message) error {
	if msg == nil {
		return errors.New("msg is nil")
	}

	if kc.producer == nil {
		return errors.New("nil Producer")
	}

	//需要序化后才能传输
	if rawData, err := json.Marshal(msg);  err != nil {
		return err
	} else {
		//发送到kafka的broker
		kc.producer.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: rawData}
	}

	return nil
}

func (kc *KafkaClient) Stop() error {
	kc.producer.Close()
	kc.consumer.Close()
	return nil
}

func Uint16ToBytes(n uint16) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
	}
}

func BytesToUint16(array []byte) uint16 {
	var data uint16 = 0
	for i := 0; i < len(array); i++ {
		data = data + uint16(uint(array[i])<<uint(8*i))
	}

	return data
}

func Uint32ToBytes(n uint32) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
		byte(n >> 16),
		byte(n >> 24),
	}
}

var ProviderSet = wire.NewSet(NewKafkaOptions, NewKafkaClient)
