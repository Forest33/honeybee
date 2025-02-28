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
				uc.subscribers.add(e.Topic, e.Script)
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
