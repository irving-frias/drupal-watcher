package orchestrator

import (
	"context"
	"time"
)

type OrchestratorService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Stats() (changes, clears int64)
	StartTime() time.Time
	Engine() *Engine
}

type orchestratorService struct {
	engine *Engine
}

func NewOrchestratorService(engine *Engine) OrchestratorService {
	return &orchestratorService{engine: engine}
}

func (s *orchestratorService) Start(ctx context.Context) error {
	return s.engine.Run(ctx)
}

func (s *orchestratorService) Stop(ctx context.Context) error {
	return nil
}

func (s *orchestratorService) Stats() (changes, clears int64) {
	return s.engine.Stats()
}

func (s *orchestratorService) StartTime() time.Time {
	return s.engine.StartTime()
}

func (s *orchestratorService) Engine() *Engine {
	return s.engine
}
