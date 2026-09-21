package services

import (
	"sync"
	"testing"

	"github.com/toboshii/hajimari/internal/models"
)

// TestGetCachedKubeAppsConcurrentAccess exercises the read path of
// GetCachedKubeApps against concurrent cache swaps by the updater logic.
// Run with -race: the detector flags the unsynchronized read of the
// package-level kubeAppCache if the RLock in GetCachedKubeApps is removed.
func TestGetCachedKubeAppsConcurrentAccess(t *testing.T) {
	svc := &appService{}

	const iterations = 5000
	var wg sync.WaitGroup

	// Writer goroutine: mimics startKubeAppCacheUpdater swapping the
	// package-level cache under mutex.Lock().
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			fresh := []models.AppGroup{{Group: "race-test"}}
			mutex.Lock()
			kubeAppCache = fresh
			mutex.Unlock()
		}
	}()

	// Reader goroutines: repeatedly call GetCachedKubeApps.
	const readers = 4
	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = svc.GetCachedKubeApps()
			}
		}()
	}

	wg.Wait()

	// Sanity check: the cache is still readable and non-nil after the swaps.
	mutex.RLock()
	cache := kubeAppCache
	mutex.RUnlock()
	if cache == nil {
		t.Fatal("kubeAppCache should not be nil after concurrent updates")
	}
	if len(cache) != 1 || cache[0].Group != "race-test" {
		t.Fatalf("unexpected kubeAppCache contents: %v", cache)
	}
}
