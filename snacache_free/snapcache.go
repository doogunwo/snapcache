package snapcache_free

import (
	"container/list"
	"sync"
	"sync/atomic"
	"unsafe"
)

type SnapCache[K comparable, V any] struct {
	main unsafe.Pointer   // 원자적으로 접근하기 위한 리스트 포인터
	pointer config        // 캐시 동작에 필요한 설정값들
	maxSize int64         // 최대 크기
	currentSize int64     // 현재 크기 (원자적으로 접근)
	items sync.Map        // 원자적 맵
	max int64             // 최대 플래그 값 (CAS)
}

type config struct {
	cushioning int64
	snap int64
	mid int64
}

type entry[K comparable, V any] struct {
	key K
	value V
	element unsafe.Pointer  // 리스트 엘리먼트를 원자적으로 저장
	flag    int64           // CAS에 사용되는 플래그
}

func New[K comparable, V any](maxSize int64) *SnapCache[K, V] {
	mid := int64(516)
	snap := int64(516)
	cushioning := int64(0)

	mainList := list.New()
	return &SnapCache[K, V]{
		main:        unsafe.Pointer(mainList),
		pointer:     config{mid: mid, snap: snap, cushioning: cushioning},
		maxSize:     maxSize,
		currentSize: 0,
	}
}

func (sc *SnapCache[K, V]) Set(key K, value V) {
	main := (*list.List)(atomic.LoadPointer(&sc.main))

	// 원자적 맵 접근
	loadedEntry, ok := sc.items.Load(key)
	if ok {
		// 기존 값이 있으면 업데이트
		e := loadedEntry.(*entry[K, V])
		e.value = value
		return
	}

	// 캐시가 가득 차면 축출
	if atomic.LoadInt64(&sc.currentSize) >= sc.maxSize {
		sc.Evict()
	}

	e := &entry[K, V]{
		key:     key,
		value:   value,
		flag:    int64(0),
		element: unsafe.Pointer(main.PushFront(e)), // 리스트에 원자적으로 엘리먼트 추가
	}

	sc.items.Store(key, e)                       // 원자적으로 아이템 저장
	atomic.AddInt64(&sc.currentSize, 1)           // 현재 크기 업데이트
}

func (sc *SnapCache[K, V]) Evict() int {
	evictCounter := 0
	evictSize := sc.pointer.snap - sc.pointer.cushioning

	for atomic.LoadInt64(&sc.currentSize) > 0 && evictSize > 0 {
		main := (*list.List)(atomic.LoadPointer(&sc.main))

		front := main.Front()
		if front == nil {
			break
		}

		e := front.Value.(*entry[K, V])

		if atomic.LoadInt64(&e.flag) >= atomic.LoadInt64(&sc.max) {
			main.Remove(front)
			atomic.StorePointer(&e.element, unsafe.Pointer(main.PushBack(e.key)))
		} else {
			main.Remove(front)
			sc.items.Delete(e.key)                  // 원자적으로 삭제
			atomic.AddInt64(&sc.currentSize, -1)    // 원자적으로 크기 감소
			evictCounter++
			evictSize--
		}
	}

	atomic.StoreInt64(&sc.pointer.mid, sc.maxSize/2)
	atomic.StoreInt64(&sc.pointer.cushioning, 0)

	return evictCounter
}

func (sc *SnapCache[K, V]) Get(key K) (V, bool) {
	// 원자적 맵 접근
	loadedEntry, ok := sc.items.Load(key)
	if !ok {
		atomic.AddInt64(&sc.pointer.mid, -1)
		var zeroValue V
		return zeroValue, false
	}

	e := loadedEntry.(*entry[K, V])

	// 플래그 원자적 증가
	atomic.AddInt64(&e.flag, 1)
	if e.flag > sc.max {
		atomic.StoreInt64(&sc.max, e.flag)
	}

	atomic.AddInt64(&sc.pointer.cushioning, 1)
	return e.value, true
}

// 캐시를 초기화
func (sc *SnapCache[K, V]) Purge() {
	main := (*list.List)(atomic.LoadPointer(&sc.main))
	main.Init()

	// 원자적으로 맵과 크기 초기화
	sc.items = sync.Map{}
	atomic.StoreInt64(&sc.currentSize, 0)
	atomic.StoreInt64(&sc.pointer.cushioning, 0)
	atomic.StoreInt64(&sc.max, 0)
}
