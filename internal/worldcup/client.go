//go:build worldcup

package worldcup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	baseURL        = "https://worldcup26.ir"
	cacheTTL       = 60 * time.Second
	envFeature     = "DRUPAL_WATCHER_WORLDCUP"
	tournamentEnd  = "2026-07-19"
)

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

type Client struct {
	http  *http.Client
	mu    sync.RWMutex
	cache map[string]*cacheEntry
	Teams map[string]Team
}

func NewClient() *Client {
	return &Client{
		http:  &http.Client{Timeout: 10 * time.Second},
		cache: make(map[string]*cacheEntry),
	}
}

func Enabled() bool {
	if os.Getenv(envFeature) != "1" {
		return false
	}
	if time.Now().After(tournamentEndDate()) {
		return false
	}
	return true
}

func tournamentEndDate() time.Time {
	t, err := time.Parse("2006-01-02", tournamentEnd)
	if err != nil {
		return time.Date(2026, 7, 19, 23, 59, 59, 0, time.UTC)
	}
	return t.Add(24 * time.Hour)
}

func (c *Client) get(path string, dst interface{}) error {
	c.mu.RLock()
	entry, ok := c.cache[path]
	c.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		return json.Unmarshal(entry.data.([]byte), dst)
	}

	resp, err := c.http.Get(baseURL + path)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	c.mu.Lock()
	c.cache[path] = &cacheEntry{data: body, expiresAt: time.Now().Add(cacheTTL)}
	c.mu.Unlock()

	return json.Unmarshal(body, dst)
}

func (c *Client) FetchGames() ([]Game, error) {
	var resp GamesResponse
	if err := c.get("/get/games", &resp); err != nil {
		return nil, err
	}
	return resp.Games, nil
}

func (c *Client) FetchGroups() ([]Group, error) {
	var resp GroupsResponse
	if err := c.get("/get/groups", &resp); err != nil {
		return nil, err
	}
	return resp.Groups, nil
}

func (c *Client) FetchTeams() ([]Team, error) {
	var resp TeamsResponse
	if err := c.get("/get/teams", &resp); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.Teams = make(map[string]Team)
	for _, t := range resp.Teams {
		c.Teams[t.ID] = t
	}
	c.mu.Unlock()

	return resp.Teams, nil
}

func (c *Client) EnsureTeams() error {
	c.mu.RLock()
	_, ok := c.Teams["1"]
	c.mu.RUnlock()
	if ok {
		return nil
	}
	_, err := c.FetchTeams()
	return err
}

func (c *Client) TeamName(id string) string {
	c.mu.RLock()
	t, ok := c.Teams[id]
	c.mu.RUnlock()
	if ok {
		return t.NameEN
	}
	return id
}

func (c *Client) InvalidateCache() {
	c.mu.Lock()
	c.cache = make(map[string]*cacheEntry)
	c.mu.Unlock()
}
