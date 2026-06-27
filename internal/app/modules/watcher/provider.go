package watcher

type ConfigProvider interface {
	GetRoutes() []string
	GetExcludePatterns() []string
}
