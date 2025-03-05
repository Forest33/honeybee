package usecase

import (
	"context"
	"sync"

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

	wgConnect := &sync.WaitGroup{}
	wgConnect.Add(1)

	uc.mqtt.SetConnectHandler(func() { wgConnect.Done() })
	if err := uc.mqtt.Connect(); err != nil {
		return nil, err
	}

	wgConnect.Wait()

	if err := uc.sh.Start(); err != nil {
		return nil, err
	}

	return uc, nil
}

func (uc *ScriptUseCase) mqttMessage(m entity.MQTTMessage) {
	uc.log.Debug().Str("topic", m.Topic()).Str("payload", string(m.Payload())).Msg("MQTT message")

	scripts := uc.subscribers.getScriptsByTopic(m.Topic())
	if len(scripts) == 0 {
		return
	}

	uc.sh.SendMessageEvent(scripts, m.Topic(), m.Data())
}
