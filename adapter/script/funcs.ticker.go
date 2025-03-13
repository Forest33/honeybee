package script

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (s *Script) createFnNewTicker(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.ToString(1)
		interval := L.ToInt64(2)
		data := L.ToTable(3)

		if len(name) == 0 || interval == 0 {
			s.log.Error().
				Str("script", sc.path).
				Str("name", name).
				Int64("interval", interval).
				Msg("newTicker incorrect arguments")
			return 0
		}

		t := sc.createTicker(name, time.Duration(interval))
		if t == nil {
			return 0
		}

		go func() {
			defer sc.deleteTicker(name)

			for {
				select {
				case <-t.ctx.Done():
					s.log.Debug().Str("script", sc.path).Str("name", name).Msg("ticker finished")
					return
				case <-t.t.C:
					fn := L.GetGlobal(scriptFuncOnTicker)
					if fn == lua.LNil || fn == nil || sc.state == nil {
						s.log.Warn().Str("script", sc.path).Msg("OnTicker function not found, resetting timer")
						return
					}
					if err := L.CallByParam(lua.P{
						Fn:   fn,
						NRet: 0,
					}, lua.LString(name), data); err != nil {
						s.log.Error().Err(err).Str("script", sc.path).Msg("failed to call OnTicker function")
					}
				}
			}
		}()

		return 0
	}
}

func (s *Script) createFnStopTicker(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.ToString(1)
		if len(name) == 0 {
			s.log.Error().Str("script", sc.path).Msg("stopTicker incorrect arguments")
			L.Push(lua.LBool(false))
			return 1
		}

		if !sc.deleteTicker(name) {
			s.log.Error().Str("script", sc.path).Str("name", name).Msg("ticker not found")
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(true))

		return 1
	}
}
