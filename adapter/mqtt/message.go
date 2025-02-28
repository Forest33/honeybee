package mqtt

import (
	"github.com/forest33/honeybee/pkg/codec"
)

type message struct {
	topic   string
	payload []byte
	codec   codec.Codec
	data    map[string]interface{}
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) Payload() []byte {
	return m.payload
}

func (m *message) Data() map[string]interface{} {
	return m.data
}

func (c *Client) newMessage(topic string, payload []byte) (*message, error) {
	var (
		data map[string]interface{}
	)

	if err := c.codec.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	return &message{
		topic:   topic,
		payload: payload,
		codec:   c.codec,
		data:    data,
	}, nil
}
