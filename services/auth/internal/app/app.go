package app

import "context"

type App struct {
	c *Container
}

func New(c *Container) *App {
	return &App{c: c}
}

func (a *App) Run(ctx context.Context) error {
	return runLifecycle(ctx, a.c)
}
