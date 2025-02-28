package usecase

import (
	"context"

	"github.com/forest33/honeybee/business/entity"
	"github.com/forest33/honeybee/pkg/logger"
)

const (
	eventsChannelCapacity = 100
)

type ScriptUseCase struct {
	ctx         context.Context
	cfg         *entity.Config
	log         *logger.Logger
	mqtt        MqttClient
	sh          ScriptHandler
	subscribeCh chan *entity.SubscribeEvent
	publishCh   chan *entity.PublishEvent
	subscribers *subscribers
}

func NewScriptUseCase(ctx context.Context, cfg *entity.Config, log *logger.Logger, mqtt MqttClient, sh ScriptHandler, bot entity.BotHandler, notify entity.NotificationHandler) (*ScriptUseCase, error) {
	uc := &ScriptUseCase{
		ctx:         ctx,
		cfg:         cfg,
		log:         log,
		mqtt:        mqtt,
		sh:          sh,
		subscribeCh: make(chan *entity.SubscribeEvent, eventsChannelCapacity),
		publishCh:   make(chan *entity.PublishEvent, eventsChannelCapacity),
		subscribers: newSubscribers(),
	}

	uc.sh.SetSubscribeChannel(uc.subscribeCh)
	uc.sh.SetPublishChannel(uc.publishCh)
	uc.sh.SetBotHandler(bot)
	uc.sh.SetNotificationHandler(notify)

	uc.subscribeEventHandler()
	uc.publishEventHandler()

	if err := uc.sh.Start(); err != nil {
		return nil, err
	}

	uc.mqtt.SetConnectHandler(uc.OnConnect)
	if err := uc.mqtt.Connect(); err != nil {
		return nil, err
	}

	return uc, nil
}

func (uc *ScriptUseCase) OnConnect() {
	for _, t := range uc.subscribers.getTopics() {
		if err := uc.mqtt.Subscribe(t, uc.mqttMessage); err != nil {
			uc.log.Fatalf("failed to subscribe to topic %s: %v", t, err)
		}
		uc.log.Info().Str("topic", t).Msg("subscribed to topic")
	}
}

func (uc *ScriptUseCase) mqttMessage(m entity.MQTTMessage) {
	uc.log.Debug().Str("topic", m.Topic()).Str("payload", string(m.Payload())).Msg("MQTT message")

	scripts := uc.subscribers.getScriptsByTopic(m.Topic())
	if len(scripts) == 0 {
		return
	}

	uc.sh.SendMessageEvent(scripts, m.Topic(), m.Data())
}
