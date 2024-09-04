package main

import (
	"testing"
	"strconv"
	"hash/fnv"
)

func TestSnapCache_New(t *testing.T){

	cache := New[string, int](512)	
	cache.Set("key1",1)
	value, ok := cache.Get("key1")
	if !ok {
        t.Errorf("Failed to get value for key1")
    } else {
        t.Logf("Value for key1: %d", value)
    }
}

func TestSnapCache_Hit(t *testing.T){
	cache := New[string, int](512)
	// 캐시를 최대 용량까지 채웁니다
    for i := 0; i < cache.maxSize; i++ {
        cache.Set(strconv.Itoa(i), i)
    }

	for i := 0; i < 5; i++ {
        cache.Get(strconv.Itoa(i))
    }

}

func TestSnapCache_Miss(t *testing.T){
	cache := New[string, int](512)
	// 캐시를 최대 용량까지 채웁니다
    for i := 0; i < cache.maxSize; i++ {
        cache.Set(strconv.Itoa(i), i)
    }
	//512
	_, ok := cache.Get("513")
	if ok {
		t.Error("Not find key , Value")
	}
	//128
	
}

func TestSnapCache_Evict_Miss(t *testing.T){
	cache := New[string, int](512)
	// 캐시를 최대 용량까지 채웁니다
    for i := 0; i < cache.maxSize; i++ {
        cache.Set(strconv.Itoa(i), i)
    }
	t.Log(cache.currentSize)
	_, ok := cache.Get("513")
	if ok {
		t.Error("Not find key , Value")
	}
	//64
	_, ok = cache.Get("513")
	if ok {
		t.Error("Not find key , Value")
	}
	//32
	_, ok = cache.Get("513")
	if ok {
		t.Error("Not find key , Value")
	}
	//16
	_, ok = cache.Get("513")
	if ok {
		t.Error("Not find key , Value")
	}
	if cache.pointer.snap == cache.pointer.mid {
		cache.Evict()
	}
}

func TestSnapCache_Evict_Full(t *testing.T){
	cache := New[string, int](512)
	
	for i := 0; i < 700; i++ {
		cache.Set(strconv.Itoa(i), i)
	}
}

// FNV-1a 해시 함수 구현 (32비트)
func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func TestSnapCache_Key_Hash(t *testing.T){
	cache := New[string, int](512)

	for i := 0; i < cache.maxSize; i++ {
		hashedKey := hashKey(strconv.Itoa(i))
        cache.Set(strconv.Itoa(int(hashedKey)), i)
    }
}

func TestSnapCache_flage(t *testing.T){
	cache := New[string, int](512)
		
	for i := 0; i < 512; i++ {
		cache.Set(strconv.Itoa(i), i)
	}
	_, ok := cache.Get(strconv.Itoa(0))
	if !ok {
		t.Fatalf("Key '0' should be present in cache")
	}

	cache.Set(strconv.Itoa(550), 550)

	lastElement := cache.main.Back() // main 큐의 마지막 요소를 가져옵니다
	if lastElement == nil {
		t.Fatalf("Main queue is empty after Evict")
	}

	lastEntry := lastElement.Value.(*entry[string, int])
	t.Log(lastEntry.key)

}