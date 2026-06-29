package adapters

import (
	"crypto/sha1"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/irving-frias/drupal-watcher/pkg/core"
)

const lintCacheTTL = 5 * time.Minute
const lintCacheMaxEntries = 1000

type cacheEntry struct {
	result    *core.LintResult
	timestamp time.Time
	checksum  string
}

type CachingLintChecker struct {
	inner core.LintChecker
	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

func NewCachingLintChecker(inner core.LintChecker) *CachingLintChecker {
	c := &CachingLintChecker{
		inner: inner,
		cache: make(map[string]*cacheEntry),
	}
	go c.cleanupLoop()
	return c
}

func (c *CachingLintChecker) Lint(filePath string) *core.LintResult {
	checksum, err := fileChecksum(filePath)
	if err != nil {
		return c.inner.Lint(filePath)
	}

	key := fmt.Sprintf("%s:%s", filePath, checksum)

	c.mu.RLock()
	entry, ok := c.cache[key]
	c.mu.RUnlock()

	if ok && time.Since(entry.timestamp) < lintCacheTTL {
		return entry.result
	}

	result := c.inner.Lint(filePath)

	c.mu.Lock()
	if len(c.cache) >= lintCacheMaxEntries {
		c.evictLocked()
	}
	c.cache[key] = &cacheEntry{
		result:    result,
		timestamp: time.Now(),
		checksum:  checksum,
	}
	c.mu.Unlock()

	return result
}

func (c *CachingLintChecker) Invalidate(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.cache {
		if extractPath(k) == filePath {
			delete(c.cache, k)
		}
	}
}

func (c *CachingLintChecker) cleanupLoop() {
	ticker := time.NewTicker(lintCacheTTL)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, entry := range c.cache {
			if now.Sub(entry.timestamp) > lintCacheTTL*2 {
				delete(c.cache, k)
			}
		}
		c.mu.Unlock()
	}
}

func (c *CachingLintChecker) evictLocked() {
	oldest := time.Now()
	var oldestKey string
	for k, entry := range c.cache {
		if entry.timestamp.Before(oldest) {
			oldest = entry.timestamp
			oldestKey = k
		}
	}
	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

func fileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha1.Sum(data)
	return fmt.Sprintf("%x", h), nil
}

func extractPath(key string) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return key[:i]
		}
	}
	return key
}
