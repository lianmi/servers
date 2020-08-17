package kafkaBackend

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"

	// "time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/pkg/errors"

	// uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	// Auth "github.com/lianmi/servers/api/proto/auth"
	"github.com/lianmi/servers/internal/pkg/models"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type KafkaOptions struct {
	Broker string
	Group  string
	Topics string //以逗号隔开: Auth.Frontend Users.Frontend ... etc.
}

type KafkaClient struct {
	o                 *KafkaOptions
	app               string
	topics            []string
	recvFromFrontChan chan *models.Message //接收到payload
	consumer          *kafka.Consumer
	producer          *kafka.Producer
	logger            *zap.Logger
	redisPool         *redis.Pool
	//定义key=cmdid的处理func，当收到消息后，从此map里查出对应的处理方法
	handleFuncMap map[uint32]func(payload *models.Message) error
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

func NewKafkaClient(o *KafkaOptions, redisPool *redis.Pool, logger *zap.Logger) *KafkaClient {
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
	kClient := &KafkaClient{
		o:                 o,
		topics:            topics,
		recvFromFrontChan: make(chan *models.Message, 10),
		consumer:          c,
		producer:          p,
		logger:            logger.With(zap.String("type", "kafka.Client")),
		redisPool:         redisPool,
		handleFuncMap:     make(map[uint32]func(payload *models.Message) error),
	}

	//注册每个业务子类型的处理方法

	return kClient
}

func (kc *KafkaClient) Application(name string) {
	kc.app = name
}

//启动Kafka实例
func (kc *KafkaClient) Start() error {
	kc.logger.Info("==> Subscribe Topics ", zap.Strings("Topics", kc.topics))

	err := kc.consumer.SubscribeTopics(kc.topics, nil)
	// err := kc.consumer.SubscribeTopics([]string{"auth.Backend"}, nil)
	if err != nil {
		kc.logger.Error("Failed to SubscribeTopics: ", zap.Error(err))
		return err
	}

	//尝试读取redis
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()
	// vkey := fmt.Sprintf("verificationCode:%s", email)

	if bar, err := redis.String(redisConn.Do("GET", "bar")); err == nil {
		kc.logger.Info("redisConn GET ", zap.String("bar", bar))
	}

	//Go程，处理dispatcher发来的业务数据
	go kc.ProcessRecvPayload()

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

			case ev := <-kc.consumer.Events():
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					// kc.logger.Info("AssignedPartitions: ", zap.String("e:", e))
					kc.consumer.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					// kc.logger.Info("RevokedPartitions: ", zap.String("e:", e))
					kc.consumer.Unassign()
				case *kafka.Message:
					kc.logger.Info("Message on: ", zap.String("TopicPartition:", e.TopicPartition.String()))

					kfaPayload := new(models.Message)

					if err := json.Unmarshal(e.Value, kfaPayload); err == nil {
						kc.recvFromFrontChan <- kfaPayload //将来自dispatcher的数据压入本地通道

					} else {
						kc.logger.Error("json.Unmarshal Error", zap.Error(err))
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

// 处理dispatcher发来的业务数据
func (kc *KafkaClient) ProcessRecvPayload() {
	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			kc.logger.Info("Caught signal terminating")
			_ = sig
			run = false
		case msg := <-kc.recvFromFrontChan: //读取来着dispatcher的数据
			taskId := msg.GetTaskID()
			businessType := uint16(msg.GetBusinessType())
			businessSubType := uint16(msg.GetBusinessSubType())
			businessTypeName := msg.GetBusinessTypeName()

			//根据目标target,  组装数据包， 写入processChan
			kc.logger.Info("kfaPayload",
				// zap.String("Topic:", payload.Topic),
				zap.Uint32("taskId:", taskId),                     //taskId
				zap.String("BusinessTypeName:", businessTypeName), //业务名称
				zap.Uint16("businessType:", businessType),         // 业务类型
				zap.Uint16("businessSubType:", businessSubType),   // 业务子类型
				zap.String("Source:", msg.GetSource()),            // 业务数据发送者, 这里是businessTypeName
				zap.String("Target:", msg.GetTarget()),            // 接收者, 这里是自己，authService
			)

			//根据businessType以及businessSubType进行处理, func
			// var ok bool
			if handleFunc, ok := kc.handleFuncMap[UnionUint16ToUint32(businessType, businessSubType)]; !ok {
				kc.logger.Warn("Can not process this businessType", zap.Uint16("businessType:", businessType), zap.Uint16("businessSubType:", businessSubType))
				continue
			} else {
				if err := handleFunc(msg); err != nil {

					msg.SetCode(500) //异常出错
					msg.SetErrorMsg([]byte("Internal Server Error"))

					//处理完成，向dispatcher发送
					topic := msg.GetSource() + ".Frontend"
					kc.Produce(topic, msg)
				}
			}

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
	rawData, _ := json.Marshal(msg)

	kc.producer.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: rawData}

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

func BytesToUint32(buf []byte) uint32 {
	b_buf := bytes.NewBuffer(buf)
	var x uint32
	binary.Read(b_buf, binary.BigEndian, &x)
	return x
}

//将两个uint16的数字合并为一个uint32
func UnionUint16ToUint32(a uint16, b uint16) uint32 {
	a_buf := Uint16ToBytes(a)
	b_buf := Uint16ToBytes(b)

	a_buf = append(a_buf, b_buf[:]...)
	return BytesToUint32(a_buf)
}

var ProviderSet = wire.NewSet(NewKafkaOptions, NewKafkaClient)
