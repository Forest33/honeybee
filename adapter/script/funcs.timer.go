package script

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (s *Script) createFnNewTimer(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.ToString(1)
		delay := L.ToInt64(2)
		data := L.ToTable(3)

		if len(name) == 0 || delay == 0 {
			s.log.Error().
				Str("script", sc.path).
				Str("name", name).
				Int64("delay", delay).
				Msg("newTimer incorrect arguments")
			return 0
		}

		t := sc.createTimer(name, time.Duration(delay))
		if t == nil {
			return 0
		}

		go func() {
			defer sc.deleteTimer(name)

			select {
			case <-t.ctx.Done():
				s.log.Debug().Str("script", sc.path).Str("name", name).Msg("timer finished")
				return
			case <-t.t.C:
				fn := L.GetGlobal(scriptFuncOnTimer)
				if fn == lua.LNil || fn == nil || sc.state == nil {
					s.log.Warn().Str("script", sc.path).Msg("OnTimer function not found, resetting timer")
					t.t.Reset(time.Duration(delay))
					return
				}
				if err := L.CallByParam(lua.P{
					Fn:   fn,
					NRet: 0,
				}, lua.LString(name), data); err != nil {
					s.log.Error().Err(err).Str("script", sc.path).Msg("failed to call OnTimer function")
				}
			}
		}()

		return 0
	}
}

func (s *Script) createFnStopTimer(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.ToString(1)
		if len(name) == 0 {
			s.log.Error().Str("script", sc.path).Msg("stopTimer incorrect arguments")
		}

		if !sc.deleteTimer(name) {
			s.log.Error().Str("script", sc.path).Str("name", name).Msg("timer not found")
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(true))

		return 1
	}
}
