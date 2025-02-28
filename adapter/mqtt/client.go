package mqtt

import (
	"context"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/forest33/honeybee/business/entity"
	"github.com/forest33/honeybee/pkg/codec"
	"github.com/forest33/honeybee/pkg/logger"
)

type Client struct {
	cfg                       *Config
	log                       *logger.Logger
	cli                       mqtt.Client
	codec                     codec.Codec
	externalConnectHandler    ConnectHandler
	externalDisconnectHandler DisconnectHandler
	subscriptions             *sync.Map
}

type MessageHandler func(m entity.MQTTMessage)
type ConnectHandler func()
type DisconnectHandler func()

func New(ctx context.Context, cfg *Config, log *logger.Logger, codec codec.Codec) (*Client, error) {
	m := &Client{
		cfg:           cfg,
		log:           log,
		codec:         codec,
		subscriptions: &sync.Map{},
	}

	tlsConfig, err := cfg.getTLSConfig()
	if err != nil {
		return nil, err
	}

	opts := mqtt.NewClientOptions()
	if !cfg.ServerTLS {
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Host, cfg.Port))
	} else {
		opts.AddBroker(fmt.Sprintf("ssl://%s:%d", cfg.Host, cfg.Port))
	}
	opts.SetClientID(fmt.Sprintf("%s-%d", cfg.ClientID, time.Now().Unix()))
	opts.SetUsername(cfg.User)
	opts.SetPassword(cfg.Password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(cfg.ConnectRetryInterval)
	opts.SetDefaultPublishHandler(m.messagePubHandler)
	opts.SetTLSConfig(tlsConfig)
	opts.OnConnect = m.connectHandler
	opts.OnConnectionLost = m.connectLostHandler
	m.cli = mqtt.NewClient(opts)

	entity.GetWg(ctx).Add(1)
	go func() {
		<-ctx.Done()
		m.Close()
		log.Info().Msg("MQTT client disconnected")
		entity.GetWg(ctx).Done()
	}()

	return m, nil
}

func (c *Client) Publish(topic string, payload []byte) error {
	token := c.cli.Publish(topic, 0, false, payload)
	if token.WaitTimeout(c.cfg.Timeout) && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *Client) Subscribe(topic string, handler MessageHandler) error {
	_, exists := c.subscriptions.LoadOrStore(topic, struct{}{})
	if exists {
		return nil
	}

	token := c.cli.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		m, err := c.newMessage(topic, msg.Payload())
		if err != nil {
			c.log.Error().Err(err).Str("topic", topic).Str("payload", string(msg.Payload())).Msg("failed to create message")
			return
		}
		handler(m)
	})
	if token.WaitTimeout(c.cfg.Timeout) && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *Client) Connect() error {
	if token := c.cli.Connect(); token.WaitTimeout(c.cfg.Timeout) && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *Client) Close() {
	c.cli.Disconnect(1000)
}

func (c *Client) SetConnectHandler(h ConnectHandler) {
	c.externalConnectHandler = h
}

func (c *Client) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	c.log.Debug().Msgf("MQTT received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

func (c *Client) connectHandler(client mqtt.Client) {
	c.log.Info().Str("host", c.cfg.Host).Int("port", c.cfg.Port).Msg("MQTT connected")
	if c.externalConnectHandler != nil {
		c.externalConnectHandler()
	}
}

func (c *Client) connectLostHandler(client mqtt.Client, err error) {
	c.log.Error().Msgf("MQTT connect lost: %v", err)
	c.subscriptions.Clear()
}
