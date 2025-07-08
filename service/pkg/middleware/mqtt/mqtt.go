package mqtt

import (
	"bytes"
	"context"
	"fmt"

	m "github.com/AliwareMQ/mqtt-server-sdk/go/server-sdk"
)

// TODO: 开源 mqtt 消息云端使用 share 模式。需要代码开发兼容。

var (
	producers []*m.ServerProducer
	consumers []*m.ServerConsumer
)

type Config struct {
	AccessKey  string
	SecretKey  string
	InstanceID string
	Domain     string
	Port       int16
	Topic      string
}

type MessageProcessor struct {
	handle func(msgID string, messageProperties *m.MessageProperties, body []byte) error
}

func (t *MessageProcessor) Process(msgID string, Properties *m.MessageProperties, body []byte) error {
	return t.handle(msgID, Properties, body)
}

type StatusProcessor struct{}

func (s *StatusProcessor) Process(statusNotice *m.StatusNotice) error {
	fmt.Printf("status: clientID=%s,channelID=%s,eventType=%s\n", statusNotice.ClientID, statusNotice.ChannelID, statusNotice.EventType)
	return nil
}

func InitPublish(_ context.Context, conf *Config) func(msg []byte) (msgID string, err error) {
	serverProducer := &m.ServerProducer{}
	channelConfig := &m.ChannelConfig{
		AccessKey:  conf.AccessKey,
		SecretKey:  conf.SecretKey,
		InstanceId: conf.InstanceID,
		Domain:     conf.Domain,
		Port:       conf.Port,
	}
	serverProducer.Start(channelConfig)
	producers = append(producers, serverProducer)

	var mqttTopic bytes.Buffer
	mqttTopic.WriteString(conf.Topic)
	mqttTopic.WriteString("/t2")

	return func(msg []byte) (msgID string, err error) {
		return serverProducer.SendMessage(mqttTopic.String(), msg)
	}
}

func InitSubscribe(_ context.Context, conf *Config, handle func(msgID string, messageProperties *m.MessageProperties, body []byte) error) error {
	serverConsumer := &m.ServerConsumer{}
	channelConfig := &m.ChannelConfig{
		AccessKey:  conf.AccessKey,
		SecretKey:  conf.SecretKey,
		InstanceId: conf.InstanceID,
		Domain:     conf.Domain,
		Port:       conf.Port,
	}
	serverConsumer.Start(channelConfig)
	consumers = append(consumers, serverConsumer)
	// 可以订阅状态通知，后面根据需要订阅
	// serverConsumer.SubscribeTopic(conf.Topic, &MessageProcessor{})
	return serverConsumer.SubscribeTopic(conf.Topic, &MessageProcessor{handle: handle})
}
