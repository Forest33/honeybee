package bot

import (
	"context"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/forest33/honeybee/pkg/logger"
	"github.com/forest33/honeybee/pkg/scheduler"
)

type Bot struct {
	ctx      context.Context
	cfg      *Config
	log      *logger.Logger
	sh       Scheduler
	bot      *tgbotapi.BotAPI
	updates  tgbotapi.UpdatesChannel
	workerCh chan string
}

type Scheduler interface {
	AddTask(t *scheduler.Task)
}

func New(ctx context.Context, cfg *Config, log *logger.Logger, sh Scheduler) (*Bot, error) {
	b := &Bot{
		ctx:      ctx,
		cfg:      cfg,
		log:      log,
		sh:       sh,
		workerCh: make(chan string, cfg.PoolSize),
	}

	if err := b.init(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bot) init() error {
	var err error

	b.bot, err = tgbotapi.NewBotAPI(b.cfg.Token)
	if err != nil {
		return err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.cfg.UpdateTimeout

	b.updates = b.bot.GetUpdatesChan(u)

	for range b.cfg.PoolSize {
		go b.worker()
	}

	b.log.Info().Msg("Telegram bot initialized")

	return nil
}

func (b *Bot) worker() {
	for {
		select {
		case <-b.ctx.Done():
			return
		case msg, ok := <-b.workerCh:
			if !ok {
				return
			}
			if err := b.sendMessage(msg); err != nil {
				b.log.Error().Err(err).Msg("failed to send message")
				if b.sh != nil {
					b.sh.AddTask(&scheduler.Task{Handler: func() error {
						return b.sendMessage(msg)
					}})
				}
			}
		}
	}
}

func (b *Bot) SendMessage(text string) {
	b.workerCh <- text
}

func (b *Bot) sendMessage(text string) error {
	if b.bot == nil {
		return errors.New("bot not initialized")
	}

	for i := range b.cfg.ChatId {
		msg := tgbotapi.NewMessage(b.cfg.ChatId[i], text)
		if _, err := b.bot.Send(msg); err != nil {
			return err
		}
	}

	return nil
}
