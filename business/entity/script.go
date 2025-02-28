package entity

type PublishEvent struct {
	Topic   string
	Payload string
}

type SubscribeEvent struct {
	Topic  string
	Script Script
}

type Script interface {
	Path() string
	Name() string
}
