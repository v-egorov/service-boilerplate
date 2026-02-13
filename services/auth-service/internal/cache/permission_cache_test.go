package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPermissionCache(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}

	cache := NewPermissionCache(cfg)
	assert.NotNil(t, cache)
}

func TestNewPermissionCache_Defaults(t *testing.T) {
	cache := NewPermissionCache(PermissionCacheConfig{})
	assert.NotNil(t, cache)

	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 10000,
	}
	cache = NewPermissionCache(cfg)
	assert.NotNil(t, cache)
}

func TestPermissionCache_GetPermissions_NotFound(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}
	cache := NewPermissionCache(cfg)

	perms, found := cache.GetPermissions("user-1")
	assert.False(t, found)
	assert.Nil(t, perms)
}

func TestPermissionCache_SetAndGetPermissions(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}
	cache := NewPermissionCache(cfg)

	permissions := []string{"objects:read", "objects:write"}
	cache.SetPermissions("user-1", permissions)

	perms, found := cache.GetPermissions("user-1")
	assert.True(t, found)
	assert.Equal(t, permissions, perms)
}

func TestPermissionCache_GetRoles_NotFound(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}
	cache := NewPermissionCache(cfg)

	roles, found := cache.GetRoles("user-1")
	assert.False(t, found)
	assert.Nil(t, roles)
}

func TestPermissionCache_SetAndGetRoles(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}
	cache := NewPermissionCache(cfg)

	roles := []string{"admin", "user"}
	cache.SetRoles("user-1", roles)

	result, found := cache.GetRoles("user-1")
	assert.True(t, found)
	assert.Equal(t, roles, result)
}

func TestPermissionCache_Invalidate(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}
	cache := NewPermissionCache(cfg)

	cache.SetPermissions("user-1", []string{"objects:read"})
	cache.SetRoles("user-1", []string{"admin"})

	cache.Invalidate("user-1")

	perms, found := cache.GetPermissions("user-1")
	assert.False(t, found)
	assert.Nil(t, perms)

	roles, found := cache.GetRoles("user-1")
	assert.False(t, found)
	assert.Nil(t, roles)
}

func TestPermissionCache_InvalidateAll(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}
	cache := NewPermissionCache(cfg)

	cache.SetPermissions("user-1", []string{"objects:read"})
	cache.SetRoles("user-1", []string{"admin"})
	cache.SetPermissions("user-2", []string{"objects:write"})
	cache.SetRoles("user-2", []string{"user"})

	cache.InvalidateAll()

	perms1, found1 := cache.GetPermissions("user-1")
	assert.False(t, found1)
	assert.Nil(t, perms1)

	perms2, found2 := cache.GetPermissions("user-2")
	assert.False(t, found2)
	assert.Nil(t, perms2)
}

func TestPermissionCache_Eviction(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 3,
	}
	cache := NewPermissionCache(cfg)

	cache.SetPermissions("user-1", []string{"perm-1"})
	cache.SetPermissions("user-2", []string{"perm-2"})
	cache.SetPermissions("user-3", []string{"perm-3"})

	_, found := cache.GetPermissions("user-1")
	assert.True(t, found)

	cache.SetPermissions("user-4", []string{"perm-4"})

	_, found = cache.GetPermissions("user-1")
	assert.False(t, found)
}

func TestPermissionCache_ConcurrentAccess(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 1000,
	}
	cache := NewPermissionCache(cfg)

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			userID := "user"
			cache.SetPermissions(userID, []string{"perm-1"})
			cache.GetPermissions(userID)
			cache.SetRoles(userID, []string{"role-1"})
			cache.GetRoles(userID)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			userID := "user"
			cache.Invalidate(userID)
		}
		done <- true
	}()

	<-done
	<-done
}

func TestPermissionCache_Size(t *testing.T) {
	cfg := PermissionCacheConfig{
		TTL:        60 * time.Second,
		MaxEntries: 100,
	}
	cache := NewPermissionCache(cfg)

	for i := 0; i < 50; i++ {
		cache.SetPermissions(
			string(rune('A'+i)),
			[]string{"perm"},
		)
	}

	size := cache.Size()
	assert.Equal(t, 50, size)
}
