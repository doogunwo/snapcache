package snapcache_free

import (
	"testing"
)

func TestSnapCache_New(t *testing.T){
	cache := New[string, int](512)	 
	cache.Set("key1",1)
	value, ok := cache.Get("key1")
	if !ok {
		t.Errorf("Failed to get value for key1")
	} else if value != 1 {
		t.Errorf("Expected value 1 for key1, got %d", value)
	}
}



