package main

import (
	"context"

	"github.com/jerobas/saas/service"
)

type App struct {
	ctx             context.Context
	Notifier        *service.Notifier
	DatabaseService *service.DatabaseService
}

func NewApp() *App {
	return &App{}
}

// startup wires operating-system integrations after Wails is ready.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Notifier = service.NewNotifier(ctx)
	if a.DatabaseService != nil {
		a.DatabaseService.SetContext(ctx)
	}
}
