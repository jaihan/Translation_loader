package unit

import (
	"context"
	"testing"
	"time"

	"github.com/wyzauto/translation-loader/internal/cache"
	"github.com/wyzauto/translation-loader/internal/domain"
	"github.com/wyzauto/translation-loader/internal/loader"
)

type fakeLoader struct{ calls int }

func (f *fakeLoader) Load(ctx context.Context, entityType string, entityIDs []string, locales []string) (map[string][]domain.Translation, error) {
	f.calls++
	out := map[string][]domain.Translation{}
	for _, id := range entityIDs {
		out[id] = []domain.Translation{{EntityType: entityType, EntityID: id, Locale: "en", FieldName: "productname", Value: "Engine Oil"}}
	}
	return out, nil
}
func (f *fakeLoader) Invalidate(entityID string) {}

func TestCachedLoaderUsesCache(t *testing.T) {
	base := &fakeLoader{}
	cached := loader.NewCachedLoader(base, cache.NewTTLCache(time.Minute))
	_, _ = cached.Load(context.Background(), "product", []string{"p1"}, []string{"en"})
	_, _ = cached.Load(context.Background(), "product", []string{"p1"}, []string{"en"})
	if base.calls != 1 {
		t.Fatalf("expected 1 call, got %d", base.calls)
	}
}

func TestInvalidateEntity(t *testing.T) {
	base := &fakeLoader{}
	cached := loader.NewCachedLoader(base, cache.NewTTLCache(time.Minute))
	_, _ = cached.Load(context.Background(), "product", []string{"p1"}, []string{"en"})
	cached.Invalidate("p1")
	_, _ = cached.Load(context.Background(), "product", []string{"p1"}, []string{"en"})
	if base.calls != 2 {
		t.Fatalf("expected 2 calls after invalidation, got %d", base.calls)
	}
}
