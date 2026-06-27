package notifications

import (
	"context"

	"github.com/irving-frias/drupal-watcher/internal/app/eventbus"
)

type NotificationService interface {
	Run(ctx context.Context, bus *eventbus.EventBus) error
}

type notificationService struct {
	providers []NotificationProvider
}

type NotificationProvider interface {
	Name() string
	Notify(ctx context.Context, topic string, event any) error
}

func NewNotificationService(providers ...NotificationProvider) NotificationService {
	return &notificationService{providers: providers}
}

func (s *notificationService) Run(ctx context.Context, bus *eventbus.EventBus) error {
	if bus == nil {
		return nil
	}
	for _, p := range s.providers {
		p := p
		bus.Subscribe(eventbus.TopicCacheClear, func(event any) {
			if err := p.Notify(ctx, eventbus.TopicCacheClear, event); err != nil {
				return
			}
		})
	}
	return nil
}
