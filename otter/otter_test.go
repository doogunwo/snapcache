package cache

import (
	"testing"
	"time"
	"github.com/maypok86/otter"
)

func BenchmarkCache_InsertRemove(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
		cache.Delete(i)
	}
}

func BenchmarkCache_Insert(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}
}

func BenchmarkCache_Lookup(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(i)
	}
}

func BenchmarkCache_AppendRemove(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(i+b.N, i)
		cache.Delete(i)
	}
}

func BenchmarkCache_Append(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(i+b.N, i)
	}
}

func BenchmarkCache_Delete(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Delete(i)
	}
}

func BenchmarkCacheParallel_Insert(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Set(b.N, b.N)
		}
	})
}

func BenchmarkCacheParallel_Lookup(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get(b.N)
		}
	})
}

func BenchmarkCacheParallel_Delete(b *testing.B) {
	cache, err := otter.MustBuilder[int, int](b.N).
		WithTTL(time.Minute).
		Build()
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Delete(b.N)
		}
	})
}

