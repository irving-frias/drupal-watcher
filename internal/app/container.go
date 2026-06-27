package app

import (
	"fmt"

	"github.com/irving-frias/drupal-watcher/internal/app/common"
)

type Container struct {
	services map[common.ServiceName]any
}

func NewContainer() *Container {
	return &Container{
		services: make(map[common.ServiceName]any),
	}
}

func (c *Container) Set(name common.ServiceName, svc any) {
	c.services[name] = svc
}

func (c *Container) Get(name common.ServiceName) any {
	return c.services[name]
}

func (c *Container) MustGet(name common.ServiceName) any {
	svc, ok := c.services[name]
	if !ok {
		panic(fmt.Sprintf("service %q not registered", name))
	}
	return svc
}
