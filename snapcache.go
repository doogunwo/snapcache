package snapcache

import (
	"container/list"
	"sync"
)

type SnapCache[K comparable, V any] struct {
	mu sync.Mutex
	main *list.List
	sub	 *list.List
	maxSize int
	items        map[K]*entry[K, V]
	max int
	snap uint64
}

type entry[K comparable, V any] struct {
	key K
	value V
	element *list.Element
}

func New[K comparable, V any](maxSize int) *SnapCache[K,V] {
	//
	snap := uint64(1)
	
	return &SnapCache[K, V]{
		main:        	list.New(),
		sub:        	list.New(),
		maxSize:     	maxSize,
		snap:			snap,
		items:       	make(map[K]*entry[K, V]),
	}
}

func (sc *SnapCache[K, V]) Full() bool {
	if sc.main.Len() >= sc.maxSize {
		return true
	}
	return false
}

func (sc *SnapCache[K, V]) Set(key K, value V) {
	sc.mu.Lock()
	
	e, ok := sc.items[key]
    if ok {
        e.value = value
		sc.mu.Unlock()
        return
    }

	if sc.Full() {
		sc.mu.Unlock()
		sc.Evict()
	}

	// 항목을 추가하기 위해 다시 잠금 설정
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	e = &entry[K, V]{
		key:     key,
		value:   value,
		element: sc.main.PushBack(e),
	}

	sc.items[key] = e
}

func (sc *SnapCache[K,V]) Evict() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	evictCounter	:= 0
	evictSize 		:= int(sc.snap)

	for sc.main.Len() > 0 && evictSize > 0 {
		front := sc.main.Front()
		if front == nil {
			break
		}

		e := front.Value.(*entry[K, V])
		sc.main.Remove(front)
        delete(sc.items, e.key)

        evictCounter++
        evictSize--
	}

	return evictCounter
} 

func (sc *SnapCache[K, V]) Get(key K) (V, bool) {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    // 메인 큐에서 항목을 조회합니다.
    e, ok := sc.items[key]
    if ok && e.element != nil {
        return e.value, true
    }

	var value V
    // 키가 존재하지 않는 경우 기본 값과 false를 반환합니다.
    return value, false
}

// Purge 메서드 추가: 캐시를 초기화
func (sc *SnapCache[K, V]) Purge() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// 캐시 항목 모두 제거
	sc.main.Init()            // 메인 리스트 초기화
	sc.items = make(map[K]*entry[K, V]) // 맵 초기화
	sc.max = 0                 // 최대 플래그 초기화
}