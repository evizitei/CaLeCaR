package cache

import (
	"errors"
)

/*Cache is the thing the server knows
how to ask about the existance of a
particular entry.  Various implementations
can be built that correspond to this interface*/
type Cache interface {
	KeyPresent(key string) bool
	GetValue(key string) (Entry, error)
	SetValue(key string, value Entry) error
}

/*NoOp is a dummy implementation.  No keys are ever present,
so it never has to replace anything.  Naive baseline.*/
type NoOp struct{}

/*KeyPresent will always be false for the no-op cache*/
func (cno *NoOp) KeyPresent(k string) bool { return false }

/*GetValue will always return an error for the no-op cache*/
func (cno *NoOp) GetValue(k string) (Entry, error) {
	return Entry{}, errors.New("Key not present")
}

/*SetValue does nothing in the no-op cache*/
func (cno *NoOp) SetValue(k string, v Entry) error { return nil }

/*useful for easily tracking the "oldest" added node in the
cache*/
type fifoNode struct {
	key   string
	entry Entry
	prev  *fifoNode
	next  *fifoNode
}

/*FiFo is a First-in-fist-out cache implementation.
When full, it will always decide to evict the oldest key added.*/
type FiFo struct {
	maxSize int
	length  int
	head    *fifoNode
	tail    *fifoNode
	lookup  map[string]*fifoNode
}

/*KeyPresent is true if the key is in the cache right now*/
func (ff *FiFo) KeyPresent(k string) bool {
	_, ok := ff.lookup[k]
	return ok
}

/*GetValue will return the entry if present in the lookup*/
func (ff *FiFo) GetValue(k string) (Entry, error) {
	node, ok := ff.lookup[k]
	if !ok {
		return Entry{}, errors.New("Key not present in lookup hash")
	}
	return node.entry, nil
}

/*SetValue inserts a new cache entry, evicting one if necessary*/
func (ff *FiFo) SetValue(k string, v Entry) error {
	if ff.length == 0 {
		// create list head/tail
		node := &fifoNode{entry: v, key: k}
		ff.head = node
		ff.tail = node
		ff.lookup[k] = node
		ff.length = 1
		return nil
	} else if ff.length == ff.maxSize {
		// evict one entry
		newNode := &fifoNode{entry: v, key: k}
		prevHead := ff.head
		delete(ff.lookup, prevHead.key)
		newHead := prevHead.next
		newHead.prev = nil
		ff.head = newHead
		prevTail := ff.tail
		prevTail.next = newNode
		newNode.prev = prevTail
		ff.tail = newNode
		ff.lookup[k] = newNode
		// length does not change
		return nil
	}
	// just grow the list
	newNode := &fifoNode{entry: v, key: k}
	prevTail := ff.tail
	prevTail.next = newNode
	newNode.prev = prevTail
	ff.tail = newNode
	ff.lookup[k] = newNode
	ff.length = ff.length + 1
	return nil
}

func newFifo(size int) *FiFo {
	lk := make(map[string]*fifoNode)
	return &FiFo{maxSize: size, length: 0, head: nil, tail: nil, lookup: lk}
}

/*useful for easily tracking the "least recently accessed" added node in the
cache*/
type lruNode struct {
	key   string
	entry Entry
	prev  *lruNode
	next  *lruNode
}

/*Lru is a cache implementation adapting to access time.
When full, it will always decide to evict the key touched the longest ago.*/
type Lru struct {
	maxSize int
	length  int
	head    *lruNode
	tail    *lruNode
	lookup  map[string]*lruNode
}

/*KeyPresent is true if the key is in the cache right now*/
func (l *Lru) KeyPresent(k string) bool {
	_, ok := l.lookup[k]
	return ok
}

/*GetValue will return the entry if present in the lookup*/
func (l *Lru) GetValue(k string) (Entry, error) {
	node, ok := l.lookup[k]
	if !ok {
		return Entry{}, errors.New("Key not present in lookup hash")
	}
	// promote entry to most recently accessed
	if node == l.tail {
		// do nothing, it's already most recently accessed
	} else if node == l.head {
		// just move head to tail
		newHead := node.next
		newHead.prev = nil
		l.head = newHead
		prevTail := l.tail
		node.prev = prevTail
		prevTail.next = node
		l.tail = node
		node.next = nil
	} else {
		// in the middle, stitch two nodes together and move to tail
		oldPrev := node.prev
		oldNext := node.next
		oldPrev.next = oldNext
		oldNext.prev = oldPrev
		prevTail := l.tail
		node.prev = prevTail
		node.next = nil
		prevTail.next = node
		l.tail = node
	}
	return node.entry, nil
}

/*SetValue inserts a new cache entry, evicting one if necessary*/
func (l *Lru) SetValue(k string, v Entry) error {
	if l.length == 0 {
		// create list head/tail
		node := &lruNode{entry: v, key: k}
		l.head = node
		l.tail = node
		l.lookup[k] = node
		l.length = 1
		return nil
	} else if l.length == l.maxSize {
		// evict one entry
		newNode := &lruNode{entry: v, key: k}
		prevHead := l.head
		delete(l.lookup, prevHead.key)
		newHead := prevHead.next
		newHead.prev = nil
		l.head = newHead
		prevTail := l.tail
		prevTail.next = newNode
		newNode.prev = prevTail
		l.tail = newNode
		l.lookup[k] = newNode
		// length does not change
		return nil
	}
	// just grow the list
	newNode := &lruNode{entry: v, key: k}
	prevTail := l.tail
	prevTail.next = newNode
	newNode.prev = prevTail
	l.tail = newNode
	l.lookup[k] = newNode
	l.length = l.length + 1
	return nil
}

func newLru(size int) *Lru {
	lk := make(map[string]*lruNode)
	return &Lru{maxSize: size, length: 0, head: nil, tail: nil, lookup: lk}
}

/*useful for easily tracking the "least frequently accessed" added node in the
cache*/
type lfuNode struct {
	key         string
	entry       Entry
	accessCount int
	prev        *lfuNode
	next        *lfuNode
}

/*Lfu is a cache implementation adapting to access frequency.
When full, it will always decide to evict the key touched the least number of times.*/
type Lfu struct {
	maxSize int
	length  int
	head    *lfuNode
	tail    *lfuNode
	lookup  map[string]*lfuNode
}

/*KeyPresent is true if the key is in the cache right now*/
func (l *Lfu) KeyPresent(k string) bool {
	_, ok := l.lookup[k]
	return ok
}

func (l *Lfu) reorderList(node *lfuNode) {
	for {
		if node.accessCount > node.next.accessCount {
			// swap positions
			if node.prev == nil {
				// node is currently HEAD
				newHead := node.next
				rightHead := newHead.next
				newHead.prev = nil
				node.next = rightHead
				node.prev = newHead
				rightHead.prev = node
				l.head = newHead
			} else {
				// node in the middle of list
				leftTail := node.prev
				swapNode := node.next
				rightHead := swapNode.next
				leftTail.next = swapNode
				swapNode.prev = leftTail
				swapNode.next = node
				node.prev = swapNode
				node.next = rightHead
				if rightHead == nil {
					// node is now tail
					l.tail = node
					return
				}
				rightHead.prev = node
			}
		} else {
			return
		}
	}
}

/*GetValue will return the entry if present in the lookup*/
func (l *Lfu) GetValue(k string) (Entry, error) {
	node, ok := l.lookup[k]
	if !ok {
		return Entry{}, errors.New("Key not present in lookup hash")
	}
	node.accessCount++
	// move node to the right until it is accessed more
	// than prev and less than next, or until it is the tail
	if node == l.tail {
		// do nothing, it's already the most frequently accessed
	} else {
		l.reorderList(node)
	}
	return node.entry, nil
}

/*SetValue inserts a new cache entry, evicting one if necessary*/
func (l *Lfu) SetValue(k string, v Entry) error {
	if l.length == 0 {
		// create list head/tail
		node := &lfuNode{entry: v, key: k, accessCount: 1}
		l.head = node
		l.tail = node
		l.lookup[k] = node
		l.length = 1
		return nil
	} else if l.length == l.maxSize {
		// evict one entry
		newNode := &lfuNode{entry: v, key: k, accessCount: 1}
		prevHead := l.head
		delete(l.lookup, prevHead.key)
		newHead := prevHead.next
		newHead.prev = nil
		l.head = newNode
		newNode.next = newHead
		newHead.prev = newNode
		l.lookup[k] = newNode
		l.reorderList(newNode)
		// length does not change
		return nil
	}
	// just grow the list
	newNode := &lfuNode{entry: v, key: k, accessCount: 1}
	oldHead := l.head
	newNode.next = oldHead
	oldHead.prev = newNode
	l.head = newNode
	l.lookup[k] = newNode
	l.reorderList(newNode)
	return nil
}

func newLfu(size int) *Lfu {
	lk := make(map[string]*lfuNode)
	return &Lfu{maxSize: size, length: 0, head: nil, tail: nil, lookup: lk}
}

/*NewCache is a factory for building a cache implementation
of the requested strategy*/
func NewCache(cacheType string, size int) (Cache, error) {
	if cacheType == "NONE" {
		return &NoOp{}, nil
	} else if cacheType == "FIFO" {
		return newFifo(size), nil
	} else if cacheType == "LRU" {
		return newLru(size), nil
	} else if cacheType == "LFU" {
		return newLfu(size), nil
	}
	return &NoOp{}, errors.New("No cache exists of type '" + cacheType + "'")
}
