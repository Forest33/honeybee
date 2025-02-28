package notification

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/forest33/honeybee/business/entity"
	"github.com/forest33/honeybee/pkg/logger"
	"github.com/forest33/honeybee/pkg/structs"
)

type Client struct {
	ctx      context.Context
	log      *logger.Logger
	cfg      *Config
	baseURL  *url.URL
	workerCh chan *message
}

type message struct {
	ctx context.Context
	*entity.NotificationMessage
}

func New(ctx context.Context, cfg *Config, log *logger.Logger) (*Client, error) {
	c := &Client{
		ctx:      ctx,
		log:      log,
		cfg:      cfg,
		workerCh: make(chan *message, cfg.PoolSize),
	}

	if err := c.init(); err != nil {
		return nil, err
	}

	c.log.Info().Msg("The notification client is initialized")

	return c, nil
}

func (c *Client) init() (err error) {
	c.baseURL, err = url.Parse(c.cfg.BaseURL)
	if err != nil {
		return
	}

	for range c.cfg.PoolSize {
		go c.worker()
	}

	return
}

func (c *Client) worker() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case m, ok := <-c.workerCh:
			if !ok {
				return
			}
			if err := c.push(m.ctx, m.NotificationMessage); err != nil {
				c.log.Error().Err(err).Msg("failed to push notification message")
			}
		}
	}
}

func (c *Client) Push(ctx context.Context, m *entity.NotificationMessage) {
	c.workerCh <- &message{
		ctx:                 ctx,
		NotificationMessage: m,
	}
}

func (c *Client) push(ctx context.Context, m *entity.NotificationMessage) error {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.JoinPath(m.Topic).String(), strings.NewReader(m.Body))
	if err != nil {
		return err
	}

	if len(m.Title) != 0 {
		req.Header.Set("Title", m.Title)
	}
	if len(m.Attach) != 0 {
		req.Header.Set("Attach", m.Attach)
	}

	req.Header.Set("Priority", structs.If(len(m.Priority) != 0, m.Priority, c.cfg.Priority))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to push notification: status %d", resp.StatusCode)
	}

	return nil
}
