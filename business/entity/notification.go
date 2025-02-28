package entity

import "context"

type NotificationMessage struct {
	Topic    string
	Title    string
	Body     string
	Priority string
	Attach   string
}

type NotificationHandler interface {
	Push(ctx context.Context, m *NotificationMessage)
}
