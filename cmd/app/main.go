package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forest33/honeybee/adapter/bot"
	"github.com/forest33/honeybee/adapter/mqtt"
	"github.com/forest33/honeybee/adapter/notification"
	"github.com/forest33/honeybee/adapter/script"
	"github.com/forest33/honeybee/business/entity"
	"github.com/forest33/honeybee/business/usecase"
	"github.com/forest33/honeybee/pkg/automaxprocs"
	"github.com/forest33/honeybee/pkg/codec"
	"github.com/forest33/honeybee/pkg/logger"
	"github.com/forest33/honeybee/pkg/scheduler"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-ctx.Done()
		cancel()
	}()

	ctx = entity.CreateWg(ctx)

	_, cfg, err := entity.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	l := logger.New(logger.Config{
		Level:             cfg.Logger.Level,
		TimeFormat:        cfg.Logger.TimeFormat,
		PrettyPrint:       cfg.Logger.PrettyPrint,
		RedirectStdLogger: cfg.Logger.RedirectStdLogger,
		DisableSampling:   cfg.Logger.DisableSampling,
		ErrorStack:        cfg.Logger.ErrorStack,
	})

	if err := automaxprocs.Init(cfg, l); err != nil {
		l.Fatal(err)
	}

	mqttClient, err := mqtt.New(ctx, &mqtt.Config{
		Host:                 cfg.MQTT.Host,
		Port:                 cfg.MQTT.Port,
		ClientID:             cfg.MQTT.ClientID,
		User:                 cfg.MQTT.User,
		Password:             cfg.MQTT.Password,
		UseTLS:               cfg.MQTT.UseTLS,
		ServerTLS:            cfg.MQTT.ServerTLS,
		CACert:               cfg.MQTT.CACert,
		Cert:                 cfg.MQTT.Cert,
		Key:                  cfg.MQTT.Key,
		InsecureSkipVerify:   false,
		ConnectRetryInterval: time.Duration(cfg.MQTT.ConnectRetryInterval) * time.Second,
		Timeout:              time.Duration(cfg.MQTT.Timeout) * time.Second,
	}, l, codec.NewFastJsonCodec())
	if err != nil {
		l.Fatal(err)
	}

	var sched *scheduler.Scheduler
	if cfg.Scheduler.Enabled {
		sched = scheduler.New(&scheduler.Config{
			MaxTasksPerSender: cfg.Scheduler.MaxTasksPerSender,
		}, l)
	}

	var tgBot *bot.Bot
	if cfg.Bot.Enabled {
		tgBot, err = bot.New(ctx, &bot.Config{
			Token:         cfg.Bot.Token,
			ChatId:        cfg.Bot.ChatId,
			UpdateTimeout: cfg.Bot.UpdateTimeout,
			PoolSize:      cfg.Bot.PoolSize,
		}, l, sched)
		if err != nil {
			l.Fatal(err)
		}
	}

	var notifyClient *notification.Client
	if cfg.Notification.Enabled {
		notifyClient, err = notification.New(ctx, &notification.Config{
			BaseURL:  cfg.Notification.BaseURL,
			Timeout:  time.Duration(cfg.Notification.Timeout) * time.Second,
			Priority: cfg.Notification.Priority,
			PoolSize: cfg.Notification.PoolSize,
		}, l, sched)
		if err != nil {
			l.Fatal(err)
		}
	}

	sh := script.New(ctx, &script.Config{
		Folder:           cfg.Scripts.Folder,
		RegistrySize:     cfg.Scripts.RegistrySize,
		RegistryMaxSize:  cfg.Scripts.RegistryMaxSize,
		RegistryGrowStep: cfg.Scripts.RegistryGrowStep,
	}, l)

	_, err = usecase.NewScriptUseCase(ctx, cfg, l, mqttClient, sh, tgBot, notifyClient)
	if err != nil {
		l.Fatal(err)
	}

	entity.GetWg(ctx).Wait()
}
