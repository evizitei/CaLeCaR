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
	return &FiFo{maxSize: size, length: 0, head: nil, lookup: lk}
}

/*NewCache is a factory for building a cache implementation
of the requested strategy*/
func NewCache(cacheType string, size int) (Cache, error) {
	if cacheType == "NONE" {
		return &NoOp{}, nil
	} else if cacheType == "FIFO" {
		return newFifo(size), nil
	}
	return &NoOp{}, errors.New("No cache exists of type '" + cacheType + "'")
}
