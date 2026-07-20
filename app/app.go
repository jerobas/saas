package main

import (
	"context"

	presentationwails "github.com/jerobas/saas/internal/presentation/wails"
)

type App struct {
	ctx             context.Context
	Notifier        *presentationwails.Notifier
	DatabaseService *presentationwails.DatabaseService
}

func NewApp() *App {
	return &App{}
}

// startup wires operating-system integrations after Wails is ready.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.Notifier = presentationwails.NewNotifier(ctx)
	if a.DatabaseService != nil {
		a.DatabaseService.SetContext(ctx)
	}
}
