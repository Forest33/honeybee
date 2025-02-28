package entity

type MQTTMessage interface {
	Topic() string
	Payload() []byte
	Data() map[string]interface{}
}
