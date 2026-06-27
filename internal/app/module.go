package app

import "context"

type Module interface {
	Name() string
	DependsOn() []Module
	Init(container *Container) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
