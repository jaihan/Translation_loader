package loader

import (
	"context"
	"sort"
	"strings"

	"github.com/wyzauto/translation-loader/internal/cache"
	"github.com/wyzauto/translation-loader/internal/domain"
)

type CachedLoader struct {
	cache *cache.TTLCache
	next  TranslationLoader
}

func NewCachedLoader(next TranslationLoader, cache *cache.TTLCache) *CachedLoader {
	return &CachedLoader{next: next, cache: cache}
}

func (l *CachedLoader) Load(ctx context.Context, entityType string, entityIDs []string, locales []string) (map[string][]domain.Translation, error) {
	result := make(map[string][]domain.Translation, len(entityIDs))
	var missing []string

	for _, id := range entityIDs {
		key := cacheKey(entityType, id, locales)
		if cached, ok := l.cache.Get(key); ok {
			result[id] = cached
			continue
		}
		missing = append(missing, id)
	}

	if len(missing) > 0 {
		loaded, err := l.next.Load(ctx, entityType, missing, locales)
		if err != nil {
			return nil, err
		}
		for id, rows := range loaded {
			result[id] = rows
			l.cache.Set(cacheKey(entityType, id, locales), id, rows)
		}
	}
	return result, nil
}

func (l *CachedLoader) Invalidate(entityID string) {
	l.cache.Invalidate(entityID)
	l.next.Invalidate(entityID)
}

func cacheKey(entityType string, entityID string, locales []string) string {
	cp := append([]string(nil), locales...)
	sort.Strings(cp)
	return entityType + ":" + entityID + ":" + strings.Join(cp, ",")
}
