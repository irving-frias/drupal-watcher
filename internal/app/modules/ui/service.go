package ui

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
)

type UIService interface {
	Run(ctx context.Context, bus *eventbus.EventBus) error
}

type uiService struct {
	provider UIProvider
}

type UIProvider interface {
	Run(ctx context.Context, bus *eventbus.EventBus) error
}

func NewUIService(provider UIProvider) UIService {
	return &uiService{provider: provider}
}

func (s *uiService) Run(ctx context.Context, bus *eventbus.EventBus) error {
	return s.provider.Run(ctx, bus)
}
