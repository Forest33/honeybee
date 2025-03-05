package usecase

func (uc *ScriptUseCase) subscribeEventHandler() {
	go func() {
		for {
			select {
			case <-uc.ctx.Done():
				return
			case e, ok := <-uc.subscribeCh:
				if !ok {
					return
				}
				uc.subscribers.add(e.Topic, e.Script, func() {
					if err := uc.mqtt.Subscribe(e.Topic, uc.mqttMessage); err != nil {
						uc.log.Fatalf("failed to subscribe to topic %s: %v", e.Topic, err)
					}
					uc.log.Info().Str("topic", e.Topic).Msg("subscribed to topic")
				})
			}
		}
	}()
}

func (uc *ScriptUseCase) publishEventHandler() {
	go func() {
		for {
			select {
			case <-uc.ctx.Done():
				return
			case e, ok := <-uc.publishCh:
				if !ok {
					return
				}
				if err := uc.mqtt.Publish(e.Topic, []byte(e.Payload)); err != nil {
					uc.log.Error().Err(err).
						Str("topic", e.Topic).
						Interface("payload", e.Payload).
						Msg("failed to publish event")
				}
			}
		}
	}()
}
