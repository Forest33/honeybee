package usecase

import (
	"github.com/forest33/honeybee/adapter/mqtt"
	"github.com/forest33/honeybee/business/entity"
)

type MqttClient interface {
	Connect() error
	Publish(topic string, payload []byte) error
	Subscribe(topic string, handler mqtt.MessageHandler) error
	SetConnectHandler(h mqtt.ConnectHandler)
	Close()
}

type ScriptHandler interface {
	Start() error
	SendMessageEvent(script []string, topic string, payload map[string]interface{})
	SetSubscribeChannel(ch chan *entity.SubscribeEvent)
	SetPublishChannel(ch chan *entity.PublishEvent)
	SetBotHandler(bot entity.BotHandler)
	SetNotificationHandler(notify entity.NotificationHandler)
}
