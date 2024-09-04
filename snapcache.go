package main

import (
	"container/list"
    "sync"
)

type SnapCache[K comparable, V any] struct {
	mu sync.Mutex
	main *list.List
	pointer config
	maxSize int
	currentSize int
	items        map[K]*entry[K, V]
	max int
}

type config struct {
	cushioning uint64
	snap uint64
	mid uint64
}

type entry[K comparable, V any] struct {
	key K
	value V
	element *list.Element
	flag    int
}

func New[K comparable, V any](maxSize int) *SnapCache[K,V] {
	//
	mid := uint64(maxSize/2)
	snap := uint64(512)
	cushioning := uint64(0)
	
	return &SnapCache[K, V]{
		main:        list.New(),
		pointer:     config{mid: mid, snap: snap, cushioning: cushioning},
		maxSize:     maxSize,
		currentSize: 0,
		items:       make(map[K]*entry[K, V]),
	}
}

func (sc *SnapCache[K, V]) Set(key K, value V) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	e, ok := sc.items[key]
    if ok {
        e.value = value
        return
    }

	if sc.currentSize >= sc.maxSize {
		sc.Evict()
	}

	e = &entry[K, V]{
		key:     key,
		value:   value,
		element: sc.main.PushFront(&entry[K, V]{key: key, value: value}),
	}

	sc.items[key] = e
	sc.currentSize++
	
}

func (sc *SnapCache[K,V]) Evict() int {
	
	evictCounter	:= 0
	evictSize 		:= int(sc.pointer.snap - sc.pointer.cushioning)

	for sc.currentSize > 0 && evictSize > 0 {
		front := sc.main.Front()
		if front == nil {
			break
		}

		e := front.Value.(*entry[K, V])

		if e.flag > sc.max {
			sc.main.Remove(front)
			e.element = sc.main.PushBack(e.key)
			
		} else {
			sc.main.Remove(front)
			delete(sc.items, e.key)
			sc.currentSize--
			evictCounter++
			evictSize--
		}

		
	}

	sc.pointer.mid = uint64(sc.maxSize / 2)
	sc.pointer.cushioning = 0

	return evictCounter
} 


func (sc *SnapCache[K, V]) Get(key K) (V, bool) {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    // 메인 큐에서 항목을 조회합니다.
    e, ok := sc.items[key]
    if ok && e.element != nil {
		e.flag++
		if e.flag > sc.max {
			sc.max = e.flag
		}
		sc.pointer.cushioning = sc.pointer.cushioning+2
        return e.value, true
    }

	sc.pointer.mid>>= 1
	var value V
    // 키가 존재하지 않는 경우 기본 값과 false를 반환합니다.
    return value, false
}
