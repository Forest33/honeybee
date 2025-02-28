package script

import (
	"context"
	"time"

	json "github.com/layeh/gopher-json"
	"github.com/yuin/gopher-lua"

	"github.com/forest33/honeybee/business/entity"
)

type timer struct {
	ctx    context.Context
	cancel context.CancelFunc
	t      *time.Timer
}

type ticker struct {
	ctx    context.Context
	cancel context.CancelFunc
	t      *time.Ticker
}

type alarm struct {
	daysOfWeek uint8
	hour       int
	minute     int
	second     int
	ctx        context.Context
	cancel     context.CancelFunc
	t          *time.Timer
}

func (t *timer) stop() bool {
	t.cancel()
	return t.t.Stop()
}

func (t *ticker) stop() {
	t.cancel()
	t.t.Stop()
}

func (s *Script) preloadFunctions(sc *script) {
	json.Preload(sc.state)

	sc.state.PreloadModule(moduleName, func(L *lua.LState) int {
		t := L.NewTable()
		L.SetFuncs(t, map[string]lua.LGFunction{
			scriptFuncPublish:     s.createFnPublish(sc),
			scriptFuncNewTimer:    s.createFnNewTimer(sc),
			scriptFuncNewTicker:   s.createFnNewTicker(sc),
			scriptFuncNewAlarm:    s.createFnNewAlarm(sc),
			scriptFuncStopTimer:   s.createFnStopTimer(sc),
			scriptFuncStopTicker:  s.createFnStopTicker(sc),
			scriptFuncStopAlarm:   s.createFnStopAlarm(sc),
			scriptFuncSendMessage: s.createFnSendMessage(sc),
			scriptFuncPushNotify:  s.createFnPushNotify(sc),
		})
		L.Push(t)
		return 1
	})
}

func (s *Script) createFnPublish(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		topic := sc.state.ToString(1)
		payload := sc.state.ToString(2)

		if len(topic) == 0 || len(payload) == 0 {
			s.log.Error().Str("script", sc.path).Str("topic", topic).Str("payload", payload).Msg("invalid topic or payload")
			return 0
		}

		s.publishCh <- &entity.PublishEvent{
			Topic:   topic,
			Payload: payload,
		}
		return 0
	}
}

func (s *Script) createFnSendMessage(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		text := sc.state.ToString(1)

		if s.bot == nil {
			s.log.Error().Str("script", sc.path).Msg("bot is not initialized")
			return 0
		}
		if len(text) == 0 {
			s.log.Error().Str("script", sc.path).Msg("empty text")
			return 0
		}

		s.bot.SendMessage(text)

		return 0
	}
}

func (s *Script) createFnPushNotify(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		topic := sc.state.ToString(1)
		title := sc.state.ToString(2)
		body := sc.state.ToString(3)
		priority := sc.state.ToString(4)
		attach := sc.state.ToString(5)

		if s.notify == nil {
			s.log.Error().Str("script", sc.path).Msg("notify is not initialized")
			return 0
		}
		if len(topic) == 0 {
			s.log.Error().Str("script", sc.path).Msg("empty topic")
			return 0
		}
		if len(body) == 0 {
			s.log.Error().Str("script", sc.path).Msg("empty body")
			return 0
		}

		s.notify.Push(sc.ctx, &entity.NotificationMessage{
			Topic:    topic,
			Title:    title,
			Body:     body,
			Priority: priority,
			Attach:   attach,
		})

		return 0
	}
}
