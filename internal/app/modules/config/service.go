package config

import (
	"github.com/irving-frias/drupal-watcher/internal/config"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type ConfigService interface {
	Load(root string) (*config.Config, error)
	Save(cfg *config.Config, root string) error
	Default(root string) config.Config
	Validate(cfg config.Config, root string) config.Config
	DetectDrupalRoot(root string) *string
	SetCustomConfigPath(path string)
	InvalidateCache(root string)

	GetRoutes(cfg *config.Config) []string
	GetPatterns(cfg *config.Config) []string
	GetCommandsPerPattern(cfg *config.Config) map[string]string
	GetResolvedSites(cfg *config.Config) []core.SiteInfo

	WritePid(root string) error
	RemovePid(root string) error
	CheckPid(root string) (interface{}, error)
	GetStarttime(root string) (int64, error)
}

type configService struct {
	manager *config.Manager
}

func NewConfigService() ConfigService {
	return &configService{manager: config.NewManager()}
}

func (s *configService) Load(root string) (*config.Config, error) {
	cfg, err := s.manager.LoadConfig(root)
	return &cfg, err
}

func (s *configService) Save(cfg *config.Config, root string) error {
	return s.manager.SaveConfig(*cfg, root)
}

func (s *configService) Default(root string) config.Config {
	return s.manager.GetDefaultConfig(root)
}

func (s *configService) Validate(cfg config.Config, root string) config.Config {
	return s.manager.ValidateConfig(cfg, root)
}

func (s *configService) DetectDrupalRoot(root string) *string {
	return s.manager.DetectDrupalRoot(root)
}

func (s *configService) SetCustomConfigPath(path string) {
	s.manager.SetCustomConfigPath(path)
}

func (s *configService) InvalidateCache(root string) {
	s.manager.InvalidateConfigCache(root)
}

func (s *configService) GetRoutes(cfg *config.Config) []string {
	return cfg.GetRoutes()
}

func (s *configService) GetPatterns(cfg *config.Config) []string {
	return cfg.GetPatterns()
}

func (s *configService) GetCommandsPerPattern(cfg *config.Config) map[string]string {
	return cfg.GetCommandsPerPattern()
}

func (s *configService) GetResolvedSites(cfg *config.Config) []core.SiteInfo {
	return cfg.GetResolvedSites()
}

func (s *configService) WritePid(root string) error {
	return config.WritePid(root)
}

func (s *configService) RemovePid(root string) error {
	return config.RemovePid(root)
}

func (s *configService) CheckPid(root string) (interface{}, error) {
	return config.CheckPid(root)
}

func (s *configService) GetStarttime(root string) (int64, error) {
	return config.GetStarttime(root)
}
