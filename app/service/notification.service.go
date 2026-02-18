package service

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Notifier struct {
	ctx context.Context
}

func NewNotifier(ctx context.Context) *Notifier {
	return &Notifier{ctx: ctx}
}

type Notification struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

func (n *Notifier) Send(title, message string) {
	notif := Notification{
		ID:        time.Now().Format("20060102150405"),
		Title:     title,
		Message:   message,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	runtime.EventsEmit(n.ctx, "notification:new", notif)
}
