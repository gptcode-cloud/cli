package intelligence

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

type CacheEntry struct {
	Backend string
	Model   string
	Reason  string
	Expires time.Time
}

type RecommendationCache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	ttl     time.Duration
}

var defaultCache *RecommendationCache
var cacheOnce sync.Once

func getCache() *RecommendationCache {
	cacheOnce.Do(func() {
		defaultCache = &RecommendationCache{
			entries: make(map[string]CacheEntry),
			ttl:     5 * time.Minute,
		}
	})
	return defaultCache
}

func cacheKey(agentType string, additionalContext string) string {
	h := sha256.New()
	h.Write([]byte(agentType))
	h.Write([]byte(additionalContext))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetCachedRecommendation(agentType string, additionalContext string) (backend string, model string, reason string, found bool) {
	cache := getCache()
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	key := cacheKey(agentType, additionalContext)
	entry, exists := cache.entries[key]

	if !exists {
		return "", "", "", false
	}

	if time.Now().After(entry.Expires) {
		return "", "", "", false
	}

	return entry.Backend, entry.Model, entry.Reason, true
}

func SetCachedRecommendation(agentType string, additionalContext string, backend string, model string, reason string) {
	cache := getCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()

	key := cacheKey(agentType, additionalContext)
	cache.entries[key] = CacheEntry{
		Backend: backend,
		Model:   model,
		Reason:  reason,
		Expires: time.Now().Add(cache.ttl),
	}

	cleanExpiredEntries(cache)
}

func cleanExpiredEntries(cache *RecommendationCache) {
	now := time.Now()
	for key, entry := range cache.entries {
		if now.After(entry.Expires) {
			delete(cache.entries, key)
		}
	}
}

func ClearCache() {
	cache := getCache()
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.entries = make(map[string]CacheEntry)
}
