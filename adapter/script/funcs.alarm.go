package script

import (
	"errors"
	"math"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (s *Script) createFnNewAlarm(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := sc.state.ToString(1)
		daysOfWeek := uint8(sc.state.ToInt(2))
		hour := sc.state.ToInt(3)
		minute := sc.state.ToInt(4)
		second := sc.state.ToInt(5)
		data := sc.state.ToTable(6)

		if len(name) == 0 {
			s.log.Error().
				Str("script", sc.path).
				Str("name", name).
				Msg("newAlarm incorrect arguments")
			return 0
		}

		if err := validateDayOfWeek(daysOfWeek); err != nil {
			s.log.Error().Str("script", sc.path).
				Str("name", name).
				Msg("newAlarm incorrect days of week")
			return 0
		}

		a := sc.createAlarm(name, daysOfWeek, hour, minute, second)
		if a == nil {
			return 0
		}

		a.start()

		go func() {
			defer sc.deleteAlarm(name)

			for {
				select {
				case <-a.ctx.Done():
					s.log.Debug().Str("script", sc.path).Str("name", name).Msg("alarm finished")
					return
				case <-a.t.C:
					fn := sc.state.GetGlobal(scriptFuncOnAlarm)
					if fn == lua.LNil || fn == nil || sc.state == nil {
						s.log.Warn().Str("script", sc.path).Msg("OnAlarm function not found, resetting timer")
						a.reset()
						return
					}
					if err := sc.state.CallByParam(lua.P{
						Fn:   fn,
						NRet: 0,
					}, lua.LString(name), data); err != nil {
						s.log.Error().Err(err).Str("script", sc.path).Msg("failed to call OnAlarm function")
					}
					a.reset()
				}
			}
		}()

		return 0
	}
}

func (s *Script) createFnStopAlarm(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := sc.state.ToString(1)
		if len(name) == 0 {
			s.log.Error().Str("script", sc.path).Msg("stopAlarm incorrect arguments")
			L.Push(lua.LBool(false))
			return 1
		}

		if !sc.deleteAlarm(name) {
			s.log.Error().Str("script", sc.path).Str("name", name).Msg("alarm not found")
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(true))

		return 1
	}
}

func (a *alarm) start() {
	a.t = time.NewTimer(a.getDelay())
}

func (a *alarm) reset() {
	a.t.Reset(a.getDelay())
}

func (a *alarm) stop() {
	a.cancel()
	a.t.Stop()
}

func (a *alarm) getDelay() time.Duration {
	var (
		execTime    time.Time
		curDate     = time.Now().UTC()
		curWeekDay  = getWeekDay(curDate)
		nextWeekDay uint8
		minDiff     float64
		minDiffDay  uint8
		d           uint8
	)

	for d = 1; d <= 7; d++ {
		if curWeekDay <= d && a.daysOfWeek&getWeekDayMask(d) != 0 {
			if a.hour > curDate.Hour() ||
				(a.hour == curDate.Hour() && a.minute > curDate.Minute()) ||
				(a.hour == curDate.Hour() && a.minute == curDate.Minute() && a.second > curDate.Second()) {
				nextWeekDay = d
				break
			}
		}
		if md := math.Abs(float64(curWeekDay - d)); md >= minDiff && a.daysOfWeek&getWeekDayMask(d) != 0 {
			minDiff = md
			minDiffDay = d
		}
	}

	if nextWeekDay != 0 {
		execTime = time.Date(curDate.Year(), curDate.Month(), curDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(nextWeekDay-curWeekDay))
	} else {
		days := int((7 + (minDiffDay - curWeekDay)) % 7)
		if minDiffDay == curWeekDay {
			days = 7
		}
		execTime = time.Date(curDate.Year(), curDate.Month(), curDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
	}

	execTime = time.Date(execTime.Year(), execTime.Month(), execTime.Day(), a.hour, a.minute, a.second, 0, time.UTC)

	return execTime.Sub(curDate)
}

func getWeekDay(t time.Time) uint8 {
	d := t.Weekday()
	if d == 0 {
		return 7
	}
	return uint8(d)
}

func getWeekDayMask(d uint8) uint8 {
	if d == 1 {
		return 1
	}
	return uint8(math.Pow(2, float64(d-1)))
}

func validateDayOfWeek(day uint8) error {
	if day < 0b00000001 || day > 0b01111111 {
		return errors.New("not a day of week")
	}
	return nil
}
