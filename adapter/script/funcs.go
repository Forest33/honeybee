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
	date       time.Time
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
		t := sc.state.NewTable()
		sc.state.SetFuncs(t, map[string]lua.LGFunction{
			scriptFuncPublish:      s.createFnPublish(sc),
			scriptFuncNewTimer:     s.createFnNewTimer(sc),
			scriptFuncNewTicker:    s.createFnNewTicker(sc),
			scriptFuncNewAlarm:     s.createFnNewAlarm(sc),
			scriptFuncStopTimer:    s.createFnStopTimer(sc),
			scriptFuncStopTicker:   s.createFnStopTicker(sc),
			scriptFuncStopAlarm:    s.createFnStopAlarm(sc),
			scriptFuncSendMessage:  s.createFnSendMessage(sc),
			scriptFuncPushNotify:   s.createFnPushNotify(sc),
			scriptFuncSetGlobal:    s.createFnSetGlobal(sc),
			scriptFuncGetGlobal:    s.createFnGetGlobal(sc),
			scriptFuncDeleteGlobal: s.createFnDeleteGlobal(sc),
		})
		sc.state.Push(t)
		return 1
	})
}

func (s *Script) createFnPublish(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		topic := L.ToString(1)
		payload := L.ToString(2)

		if len(topic) == 0 || len(payload) == 0 {
			s.log.Error().Str("script", sc.path).Str("topic", topic).Str("payload", payload).Msg("invalid topic or payload")
			return 0
		}

		s.log.Debug().
			Str("script", sc.path).
			Str("topic", topic).
			Str("payload", payload).
			Msg("publishing message")

		s.publishCh <- &entity.PublishEvent{
			Topic:   topic,
			Payload: payload,
		}
		return 0
	}
}

func (s *Script) createFnSendMessage(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		text := L.ToString(1)

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
		topic := L.ToString(1)
		title := L.ToString(2)
		body := L.ToString(3)
		priority := L.ToString(4)
		attach := L.ToString(5)

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

func (s *Script) createFnSetGlobal(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.ToString(1)
		//value := L.ToString(2)
		value := L.Get(2)

		if len(name) == 0 {
			s.log.Error().Str("script", sc.path).Str("name", name).Interface("value", value).Msg("invalid name")
			return 0
		}

		s.globalVars.Store(name, value)

		return 0
	}
}

func (s *Script) createFnGetGlobal(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.ToString(1)

		if len(name) == 0 {
			s.log.Error().Str("script", sc.path).Str("name", name).Msg("invalid name or value")
			return 0
		}

		v, ok := s.globalVars.Load(name)
		if !ok {
			L.Push(lua.LNil)
			L.Push(lua.LBool(ok))
			return 2
		}

		switch v := v.(type) {
		case lua.LBool:
			L.Push(v)
		case lua.LNumber:
			L.Push(v)
		case lua.LString:
			L.Push(v)
		case lua.LTable:
			L.Push(&v)
		default:
			s.log.Error().Str("script", sc.path).Str("name", name).Interface("value", v).Msg("invalid value type")
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}

		L.Push(lua.LBool(ok))

		return 2
	}
}

func (s *Script) createFnDeleteGlobal(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.ToString(1)
		if len(name) == 0 {
			s.log.Error().Str("script", sc.path).Msg("deleteGlobal incorrect arguments")
			L.Push(lua.LBool(false))
			return 1
		}

		_, ok := s.globalVars.LoadAndDelete(name)

		L.Push(lua.LBool(ok))

		return 1
	}
}
