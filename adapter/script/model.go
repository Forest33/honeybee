package script

import (
	"context"
	"sync"
	"time"

	"github.com/yuin/gopher-lua"
)

const (
	moduleName            = "honeybee"
	scriptFuncInit        = "Init"
	scriptFuncOnMessage   = "OnMessage"
	scriptFuncOnTimer     = "OnTimer"
	scriptFuncOnTicker    = "OnTicker"
	scriptFuncOnAlarm     = "OnAlarm"
	scriptFuncPublish     = "publish"
	scriptFuncNewTimer    = "newTimer"
	scriptFuncNewTicker   = "newTicker"
	scriptFuncNewAlarm    = "newAlarm"
	scriptFuncStopTimer   = "stopTimer"
	scriptFuncStopTicker  = "stopTicker"
	scriptFuncStopAlarm   = "stopAlarm"
	scriptFuncSendMessage = "sendMessage"
	scriptFuncPushNotify  = "pushNotify"
)

type Config struct {
	Folder           []string
	RegistrySize     int
	RegistryMaxSize  int
	RegistryGrowStep int
}

type scriptInitResponse struct {
	Name        string
	Description string
	Subscribe   []string
}

type script struct {
	name        string
	description string
	path        string
	subscribe   []string
	state       *lua.LState
	ctx         context.Context
	cancel      context.CancelFunc
	timers      *sync.Map
	tickers     *sync.Map
	alarms      *sync.Map
}

func newScript(ctx context.Context, cfg *Config, path string) *script {
	ctx, cancel := context.WithCancel(ctx)

	state := lua.NewState(lua.Options{
		RegistrySize:     cfg.RegistrySize,
		RegistryMaxSize:  cfg.RegistryMaxSize,
		RegistryGrowStep: cfg.RegistryGrowStep,
	})

	sc := &script{
		path:    path,
		state:   state,
		ctx:     ctx,
		cancel:  cancel,
		timers:  &sync.Map{},
		tickers: &sync.Map{},
		alarms:  &sync.Map{},
	}

	sc.state.SetContext(ctx)

	return sc
}

func (s *script) createTimer(name string, delay time.Duration) *timer {
	ctx, cancel := context.WithCancel(s.ctx)
	t := &timer{
		ctx:    ctx,
		cancel: cancel,
		t:      time.NewTimer(delay),
	}
	_, loaded := s.timers.LoadOrStore(name, t)
	if loaded {
		t.stop()
		return nil
	}

	return t
}

func (s *script) deleteTimer(name string) bool {
	t, loaded := s.timers.LoadAndDelete(name)
	if !loaded {
		return false
	}
	t.(*timer).stop()
	return true
}

func (s *script) createTicker(name string, interval time.Duration) *ticker {
	ctx, cancel := context.WithCancel(s.ctx)
	t := &ticker{
		ctx:    ctx,
		cancel: cancel,
		t:      time.NewTicker(interval),
	}
	_, loaded := s.tickers.LoadOrStore(name, t)
	if loaded {
		t.stop()
		return nil
	}

	return t
}

func (s *script) deleteTicker(name string) bool {
	t, loaded := s.tickers.LoadAndDelete(name)
	if !loaded {
		return false
	}
	t.(*ticker).stop()
	return true
}

func (s *script) createAlarm(name string, daysOfWeek uint8, hour, minute, second int) *alarm {
	ctx, cancel := context.WithCancel(s.ctx)
	a := &alarm{
		daysOfWeek: daysOfWeek,
		hour:       hour,
		minute:     minute,
		second:     second,
		ctx:        ctx,
		cancel:     cancel,
	}

	_, loaded := s.alarms.LoadOrStore(name, a)
	if loaded {
		cancel()
		return nil
	}

	return a
}

func (s *script) deleteAlarm(name string) bool {
	a, loaded := s.alarms.LoadAndDelete(name)
	if !loaded {
		return false
	}
	a.(*alarm).stop()
	return true
}

func (s *script) close() {
	s.cancel()
	s.state.Close()
}

func (s *script) Path() string {
	return s.path
}

func (s *script) Name() string {
	return s.name
}
