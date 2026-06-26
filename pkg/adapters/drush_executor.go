package adapters

import (
	"context"
	"strings"
	"time"

	"github.com/irving-frias/drupal-watcher/internal/drush"
	"github.com/irving-frias/drupal-watcher/pkg/core"
)

type DrushConfig interface {
	GetDrushCmd() *string
	GetDrushCommand() string
	GetDrushArgs() []string
	GetDrupalRoot() *string
	GetNotify() bool
}

type DrushExecutor struct {
	cfg DrushConfig
}

func NewDrushExecutor(cfg DrushConfig) *DrushExecutor {
	return &DrushExecutor{cfg: cfg}
}

func (e *DrushExecutor) Execute(ctx context.Context, commands []string, dir string) core.ExecutionResult {
	start := time.Now()
	result := drush.RunCacheClears(e.cfg, commands)
	return core.ExecutionResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Duration: time.Since(start),
		Command:  strings.Join(commands, " + "),
	}
}

type SiteAwareDrushExecutor struct {
	cfg      DrushConfig
	siteName string
	uri      string
}

func NewSiteAwareDrushExecutor(cfg DrushConfig, siteName, uri string) *SiteAwareDrushExecutor {
	return &SiteAwareDrushExecutor{cfg: cfg, siteName: siteName, uri: uri}
}

type siteDrushConfig struct {
	DrushConfig
	name string
	uri  string
}

func (c *siteDrushConfig) GetDrushArgs() []string {
	return append([]string{"--uri=" + c.uri}, c.DrushConfig.GetDrushArgs()...)
}

func (e *SiteAwareDrushExecutor) Execute(ctx context.Context, commands []string, dir string) core.ExecutionResult {
	siteCfg := &siteDrushConfig{
		DrushConfig: e.cfg,
		name:        e.siteName,
		uri:         e.uri,
	}
	start := time.Now()
	result := drush.RunCacheClears(siteCfg, commands)
	return core.ExecutionResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		Duration: time.Since(start),
		Command:  strings.Join(commands, " + "),
	}
}
