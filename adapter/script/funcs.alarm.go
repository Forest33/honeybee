package script

import (
	"math"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

var weekDays = map[string]uint8{
	"monday":    1,
	"tuesday":   2,
	"wednesday": 4,
	"thursday":  8,
	"friday":    16,
	"saturday":  32,
	"sunday":    64,
	"mon":       1,
	"tues":      2,
	"wed":       4,
	"thurs":     8,
	"fri":       16,
	"sat":       32,
	"sun":       64,
}

func (s *Script) createFnNewAlarm(sc *script) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := sc.state.ToString(1)
		dw := sc.state.ToTable(2)
		hour := sc.state.ToInt(3)
		minute := sc.state.ToInt(4)
		second := sc.state.ToInt(5)
		data := sc.state.ToTable(6)

		if len(name) == 0 || dw.Len() == 0 {
			s.log.Error().
				Str("script", sc.path).
				Str("name", name).
				Msg("newAlarm incorrect arguments")
			return 0
		}

		var daysOfWeek uint8
		dw.ForEach(func(_, v lua.LValue) {
			d := strings.ToLower(v.String())
			if _, ok := weekDays[d]; !ok {
				s.log.Error().Str("script", sc.path).
					Str("name", name).
					Str("day", d).
					Msg("wrong day of week")
			}
			daysOfWeek += weekDays[d]
		})

		a := sc.createAlarm(name, daysOfWeek, hour, minute, second)
		if a == nil {
			return 0
		}

		if err := a.start(); err != nil {
			s.log.Error().
				Str("script", sc.path).
				Str("name", name).
				Msg("failed to start alarm")
		}

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
						if err := a.reset(); err != nil {
							s.log.Error().Str("script", sc.path).
								Str("name", name).
								Msg("failed to reset alarm")
						}
						return
					}

					if err := sc.state.CallByParam(lua.P{
						Fn:   fn,
						NRet: 0,
					}, lua.LString(name), data); err != nil {
						s.log.Error().Err(err).Str("script", sc.path).Msg("failed to call OnAlarm function")
					}

					if err := a.reset(); err != nil {
						s.log.Error().Str("script", sc.path).
							Str("name", name).
							Msg("failed to reset alarm")
					}
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

func (a *alarm) start() error {
	d, err := a.getDelay()
	if err != nil {
		return err
	}

	a.t = time.NewTimer(d)

	return nil
}

func (a *alarm) reset() error {
	d, err := a.getDelay()
	if err != nil {
		return err
	}

	a.t.Reset(d)

	return nil
}

func (a *alarm) stop() {
	a.cancel()
	a.t.Stop()
}

func (a *alarm) getDelay() (time.Duration, error) {
	var (
		execTime    time.Time
		curDate     = time.Now()
		curWeekDay  = getWeekDay(curDate)
		nextWeekDay uint8
		minDiff     float64
		minDiffDay  uint8
		d           uint8
	)

	tz, err := time.LoadLocation("Local")
	if err != nil {
		return 0, err
	}

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
		execTime = time.Date(curDate.Year(), curDate.Month(), curDate.Day(), 0, 0, 0, 0, tz).AddDate(0, 0, int(nextWeekDay-curWeekDay))
	} else {
		days := int((7 + (minDiffDay - curWeekDay)) % 7)
		if minDiffDay == curWeekDay {
			days = 7
		}
		execTime = time.Date(curDate.Year(), curDate.Month(), curDate.Day(), 0, 0, 0, 0, tz).AddDate(0, 0, days)
	}

	execTime = time.Date(execTime.Year(), execTime.Month(), execTime.Day(), a.hour, a.minute, a.second, 0, tz)

	return execTime.Sub(curDate), nil
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
