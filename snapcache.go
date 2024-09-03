package main

import (
	"container/list"
    "sync"
	"unsafe"
)

type SnapCache[K comparable, V any] struct {
	mu sync.Mutex
	main *list.List
	sub *list.List
	pointer config
	maxSize int
	currentSize int
	items  map[interface{}]*entry
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
}

func New[K comparable, V any](maxSize int) *SnapCache {
	//
	mid := uint64(maxSize/2)
	snap := uint64(16)
	cushioning := uint64(0)
	
	return &SnapCache[K, V]{
		main:        list.New(),
		sub:         list.New(),
		pointer:     config{mid: mid, snap: snap, cushioning: cushioning},
		maxSize:     maxSize,
		currentSize: 0,
		items:       make(map[interface{}]*entry[K, V]),
	}
}

func (sc *SnapCache[K, V]) Set(key K, Value V) {

	sc.mu.Lock()
	defer sc.mu.Unlock()

	//항목이 이미 존재함
	e, ok := sc.items[key]
	if ok {
		e.value := value
		return
	}
	
	if sc.currentSize < sc.maxSize {
		e = &entry[K, V]{
			key:     key,
			value:   value,
			element: sc.main.PushFront(key),
		}

		sc.items[key] = e
		sc.currentSize++
	} else {
		e = &entry[K, V]{
			key:     key,
			value:   value,
			element: sc.sub.PushFront(key),
		}
		sc.items[key] = e
		//evict 수행
		//sub 에 있는 데이터를 main으로 옮기기 
	}
}

func (sc *SnapCache[K,V]) evict() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	evictCounter	:= 0
	evictSize 		:= int(sc.pointer.snap - sc.pointer.cushioning)
	for sc.currentSize > 0 && evictSize > 0 {
		front := sc.main.Front()
		if front == nil {
			break
		}

		e := front.Value(*entry[K,V])
		sc.main.Remove(front)
		delete(sc.items, e.key)
		sc.currentSize--
		evictCounter++
		evictSize--
	}

	return evictCounter
} 

func (sc *SnapCache[K,V]) move(evictSize int) {
	for i := 0; i < evictSize && sc.sub.Len() > 0; i++ {
        // sub 큐의 맨 앞 요소를 가져옴
        front := sc.sub.Front()
        // 해당 요소를 sub 큐에서 제거
        sc.sub.Remove(front)
        // 요소의 키를 가져옴
        key := front.Value.(K)
        // items 맵에서 항목을 찾아옴
        e, exists := sc.items[key]
        if !exists {
            continue // 만약 항목이 존재하지 않으면 다음으로 넘어감
        }

        // main 큐로 항목을 옮김
        e.element = sc.main.PushFront(key)

        // sub에서 main으로 이동한 항목을 업데이트
        sc.items[key] = e
    }
}

func (sc *SnapCache[K, V]) Get(key K) (V, bool) {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    // 메인 큐에서 항목을 조회합니다.
    e, ok := sc.items[key]
    if ok && e.element != nil {
        return e.value, true
    }

    // 키가 존재하지 않는 경우 기본 값과 false를 반환합니다.
    return nil, false
}