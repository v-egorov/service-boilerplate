package cache

import (
	"sync"
	"time"
)

type PermissionCache interface {
	GetPermissions(userID string) ([]string, bool)
	SetPermissions(userID string, permissions []string)
	GetRoles(userID string) ([]string, bool)
	SetRoles(userID string, roles []string)
	Invalidate(userID string)
	InvalidateAll()
	Size() int
}

type PermissionCacheConfig struct {
	TTL        time.Duration
	MaxEntries int
}

type permissionCache struct {
	mu          sync.RWMutex
	permissions map[string]cacheEntry
	roles       map[string]cacheEntry
	ttl         time.Duration
	maxEntries  int
}

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

func NewPermissionCache(cfg PermissionCacheConfig) PermissionCache {
	if cfg.TTL == 0 {
		cfg.TTL = 60 * time.Second
	}
	if cfg.MaxEntries == 0 {
		cfg.MaxEntries = 10000
	}

	return &permissionCache{
		permissions: make(map[string]cacheEntry),
		roles:       make(map[string]cacheEntry),
		ttl:         cfg.TTL,
		maxEntries:  cfg.MaxEntries,
	}
}

func (c *permissionCache) GetPermissions(userID string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.permissions[userID]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		return nil, false
	}

	if perms, ok := entry.data.([]string); ok {
		return perms, true
	}

	return nil, false
}

func (c *permissionCache) SetPermissions(userID string, permissions []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.permissions) >= c.maxEntries {
		c.evictOldestPermissions()
	}

	c.permissions[userID] = cacheEntry{
		data:      permissions,
		timestamp: time.Now(),
	}
}

func (c *permissionCache) GetRoles(userID string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.roles[userID]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > c.ttl {
		return nil, false
	}

	if roles, ok := entry.data.([]string); ok {
		return roles, true
	}

	return nil, false
}

func (c *permissionCache) SetRoles(userID string, roles []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.roles) >= c.maxEntries {
		c.evictOldestRoles()
	}

	c.roles[userID] = cacheEntry{
		data:      roles,
		timestamp: time.Now(),
	}
}

func (c *permissionCache) Invalidate(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.permissions, userID)
	delete(c.roles, userID)
}

func (c *permissionCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.permissions = make(map[string]cacheEntry)
	c.roles = make(map[string]cacheEntry)
}

func (c *permissionCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.permissions)
}

func (c *permissionCache) evictOldestPermissions() {
	var oldestID string
	var oldestTime time.Time

	for id, entry := range c.permissions {
		if oldestTime.IsZero() || entry.timestamp.Before(oldestTime) {
			oldestTime = entry.timestamp
			oldestID = id
		}
	}

	if oldestID != "" {
		delete(c.permissions, oldestID)
	}
}

func (c *permissionCache) evictOldestRoles() {
	var oldestID string
	var oldestTime time.Time

	for id, entry := range c.roles {
		if oldestTime.IsZero() || entry.timestamp.Before(oldestTime) {
			oldestTime = entry.timestamp
			oldestID = id
		}
	}

	if oldestID != "" {
		delete(c.roles, oldestID)
	}
}
