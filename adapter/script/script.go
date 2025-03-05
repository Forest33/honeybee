package script

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/radovskyb/watcher"
	"github.com/yuin/gluamapper"
	"github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/forest33/honeybee/business/entity"
	"github.com/forest33/honeybee/pkg/logger"
	"github.com/forest33/honeybee/pkg/structs"
)

type Script struct {
	ctx         context.Context
	cfg         *Config
	log         *logger.Logger
	scripts     *sync.Map
	watcher     *watcher.Watcher
	subscribeCh chan *entity.SubscribeEvent
	publishCh   chan *entity.PublishEvent
	bot         entity.BotHandler
	notify      entity.NotificationHandler
}

func New(ctx context.Context, cfg *Config, log *logger.Logger) *Script {
	s := &Script{
		ctx:     ctx,
		cfg:     cfg,
		log:     log,
		scripts: &sync.Map{},
	}

	entity.GetWg(ctx).Add(1)
	go func() {
		<-ctx.Done()
		log.Info().Msg("scripts handler finished")
		entity.GetWg(ctx).Done()
	}()

	return s
}

func (s *Script) SendMessageEvent(scriptPath []string, topic string, payload map[string]interface{}) {
	for i := range scriptPath {
		sc, ok := s.scripts.Load(scriptPath[i])
		if !ok {
			s.log.Error().Str("script", scriptPath[i]).Str("topic", topic).Msg("script does not exist")
			continue
		}

		if err := sc.(*script).state.CallByParam(lua.P{
			Fn:   sc.(*script).state.GetGlobal(scriptFuncOnMessage),
			NRet: 0,
		}, lua.LString(topic), luar.New(sc.(*script).state, payload)); err != nil {
			s.log.Error().Err(err).Str("script", scriptPath[i]).Str("topic", topic).Msg("failed to call OnMessage function")
		}
	}
}

func (s *Script) Start() error {
	s.initWatcher()
	return s.initScripts()
}

func (s *Script) SetSubscribeChannel(ch chan *entity.SubscribeEvent) {
	s.subscribeCh = ch
}

func (s *Script) SetPublishChannel(ch chan *entity.PublishEvent) {
	s.publishCh = ch
}

func (s *Script) SetBotHandler(bot entity.BotHandler) {
	s.bot = bot
}

func (s *Script) SetNotificationHandler(notify entity.NotificationHandler) {
	s.notify = notify
}

func (s *Script) initScripts() error {
	for _, folder := range s.cfg.Folder {
		files, err := os.ReadDir(folder)
		if err != nil {
			s.log.Error().Err(err).Str("folder", folder).Msg("error reading folder files")
			return err
		}

		if err := s.watcher.Add(folder); err != nil {
			s.log.Error().Err(err).Str("path", folder).Msg("error adding watcher")
			return err
		}

		for _, f := range files {
			path, err := filepath.Abs(filepath.Join(folder, f.Name()))
			if err != nil {
				s.log.Error().Err(err).Str("path", f.Name()).Msg("error resolving path")
			}

			if err := s.loadScript(path); err != nil {
				s.log.Error().Err(err).Str("path", path).Msg("error loading script")
				return err
			}

			if err := s.watcher.Add(path); err != nil {
				s.log.Error().Err(err).Str("path", path).Msg("error adding watcher")
				return err
			}
		}
	}

	return nil
}

func (s *Script) loadScript(path string) error {
	s.log.Debug().Str("path", path).Msg("loading script")

	sc := newScript(s.ctx, s.cfg, path)

	s.preloadFunctions(sc)

	if err := sc.state.DoFile(path); err != nil {
		return err
	}

	if err := sc.state.CallByParam(lua.P{
		Fn:   sc.state.GetGlobal(scriptFuncInit),
		NRet: 1,
	}); err != nil {
		return err
	}
	ret := sc.state.Get(-1)

	init := &scriptInitResponse{}
	if err := gluamapper.Map(ret.(*lua.LTable), init); err != nil {
		return err
	}
	sc.state.Pop(1)

	sc.subscribe = init.Subscribe
	sc.name = init.Name
	sc.description = init.Description

	s.scripts.Store(path, sc)

	structs.ForEach(init.Subscribe, func(topic string) {
		s.subscribeCh <- &entity.SubscribeEvent{
			Topic:  topic,
			Script: sc,
		}
	})

	return nil
}
