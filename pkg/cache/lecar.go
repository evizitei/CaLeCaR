package cache

import (
	"errors"
	"math"
	"math/rand"
)

type lecarLruNode struct {
	prev      *lecarLruNode
	next      *lecarLruNode
	entryNode *lecarLookupNode
}
type lecarLfuNode struct {
	prev        *lecarLfuNode
	next        *lecarLfuNode
	accessCount int
	entryNode   *lecarLookupNode
}

type lecarLookupNode struct {
	key     string
	entry   Entry
	lruNode *lecarLruNode
	lfuNode *lecarLfuNode
}

type lecarHistoryNode struct {
	key          string
	evictionType string
	next         *lecarHistoryNode
	prev         *lecarHistoryNode
}

/*Lecar balances a frequency and recency distribution
https://www.usenix.org/system/files/conference/hotstorage18/hotstorage18-paper-vietri.pdf
*/
type Lecar struct {
	maxSize       int
	length        int
	lruHead       *lecarLruNode
	lruTail       *lecarLruNode
	lfuHead       *lecarLfuNode
	lfuTail       *lecarLfuNode
	weightLru     float64
	weightLfu     float64
	lookup        map[string]*lecarLookupNode
	debug         bool
	historyLookup map[string]*lecarHistoryNode
	historyLength int
	historyHead   *lecarHistoryNode
	historyTail   *lecarHistoryNode
	lambda        float64
	discount      float64
}

func (l *Lecar) updateAlgoWeights(node *lecarHistoryNode) {
	regret := 1.0
	histPosition := 1
	pointerNode := node
	for {
		if pointerNode == l.historyTail {
			// discount is adjusted correctly now
			break
		}
		pointerNode = pointerNode.next
		regret = regret * l.discount
		histPosition = histPosition + 1
	}
	adjustCoefficient := math.Pow(math.E, (l.lambda * regret))
	wLru := l.weightLru
	wLfu := l.weightLfu
	if node.evictionType == "LRU" {
		wLfu = wLfu * adjustCoefficient
	} else if node.evictionType == "LFU" {
		wLru = wLru * adjustCoefficient
	}
	normConst := (wLfu + wLru)
	l.weightLfu = wLfu / normConst
	l.weightLru = wLru / normConst
}

/*KeyPresent is true if the key is in the cache right now*/
func (l *Lecar) KeyPresent(k string) bool {
	_, ok := l.lookup[k]
	if !ok {
		// key not in cache, check history and penalize if present
		historyNode, hOk := l.historyLookup[k]
		if hOk {
			// in history, penalize an algorithm
			l.updateAlgoWeights(historyNode)
		}
	}
	return ok
}

/*GetValue will return the entry if present in the lookup*/
func (l *Lecar) GetValue(k string) (Entry, error) {
	lookupNode, ok := l.lookup[k]
	if !ok {
		return Entry{}, errors.New("Key not present in lookup hash")
	}
	lruNode := lookupNode.lruNode
	// LRU: promote entry to most recently accessed
	if lruNode == l.lruTail {
		// do nothing, it's already most recently accessed
	} else if lruNode == l.lruHead {
		// just move head to tail
		newHead := lruNode.next
		newHead.prev = nil
		l.lruHead = newHead
		prevTail := l.lruTail
		lruNode.prev = prevTail
		prevTail.next = lruNode
		l.lruTail = lruNode
		lruNode.next = nil
	} else {
		// in the middle, stitch two nodes together and move to tail
		oldPrev := lruNode.prev
		oldNext := lruNode.next
		oldPrev.next = oldNext
		oldNext.prev = oldPrev
		prevTail := l.lruTail
		lruNode.prev = prevTail
		lruNode.next = nil
		prevTail.next = lruNode
		l.lruTail = lruNode
	}
	// LFU: increment access count and reorder
	lfuNode := lookupNode.lfuNode
	lfuNode.accessCount = lfuNode.accessCount + 1
	// move node to the right until it is accessed more
	// than prev and less than next, or until it is the tail
	if lfuNode == l.lfuTail {
		// do nothing, it's already the most frequently accessed
	} else {
		l.reorderLfuList(lfuNode)
	}
	return lookupNode.entry, nil
}

func (l *Lecar) reorderLfuList(node *lecarLfuNode) {
	for {
		if node.accessCount >= node.next.accessCount {
			// swap positions
			if node.prev == nil {
				// node is currently HEAD
				newHead := node.next
				rightHead := newHead.next
				newHead.prev = nil
				node.next = rightHead
				node.prev = newHead
				newHead.next = node
				l.lfuHead = newHead
				if rightHead == nil {
					// node is now tail
					l.lfuTail = node
					return
				}
				rightHead.prev = node
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
					l.lfuTail = node
					return
				}
				rightHead.prev = node
			}
		} else {
			return
		}
	}
}

func (l *Lecar) removeFromLru(node *lecarLruNode) {
	if node == l.lruHead {
		newLruHead := node.next
		newLruHead.prev = nil
		l.lruHead = newLruHead
		node.next = nil
	} else if node == l.lruTail {
		newLruTail := node.prev
		newLruTail.next = nil
		l.lruTail = newLruTail
		node.prev = nil
	} else {
		// node is in the middle, stitch together
		tailLeft := node.prev
		headRight := node.next
		tailLeft.next = headRight
		headRight.prev = tailLeft
		node.prev = nil
		node.next = nil
	}
}

func (l *Lecar) removeFromLfu(node *lecarLfuNode) {
	if node == l.lfuHead {
		newLfuHead := node.next
		newLfuHead.prev = nil
		l.lfuHead = newLfuHead
		node.next = nil
	} else if node == l.lfuTail {
		newLfuTail := node.prev
		newLfuTail.next = nil
		l.lfuTail = newLfuTail
		node.prev = nil
	} else {
		// node is in the middle, stitch together
		tailLeft := node.prev
		headRight := node.next
		tailLeft.next = headRight
		headRight.prev = tailLeft
		node.prev = nil
		node.next = nil
	}
}

func (l *Lecar) removeFromHistory(histNode *lecarHistoryNode) {
	if histNode == l.historyHead {
		newHistHead := histNode.next
		newHistHead.prev = nil
		l.historyHead = newHistHead
		histNode.next = nil
	} else if histNode == l.historyTail {
		newHistTail := histNode.prev
		newHistTail.next = nil
		l.historyTail = newHistTail
		histNode.prev = nil
	} else {
		// node is in the middle, stitch together
		tailLeft := histNode.prev
		headRight := histNode.next
		tailLeft.next = headRight
		headRight.prev = tailLeft
		histNode.prev = nil
		histNode.next = nil
	}
}

func (l *Lecar) appendToLfu(lfuNode *lecarLfuNode) {
	oldLfuHead := l.lfuHead
	lfuNode.next = oldLfuHead
	oldLfuHead.prev = lfuNode
	l.lfuHead = lfuNode
	l.reorderLfuList(lfuNode)
}

func (l *Lecar) appendToLru(lruNode *lecarLruNode) {
	prevLruTail := l.lruTail
	prevLruTail.next = lruNode
	lruNode.prev = prevLruTail
	l.lruTail = lruNode
}

func (l *Lecar) putInHistory(entryNode *lecarLookupNode, evictionType string) {
	historyNode := &lecarHistoryNode{key: entryNode.key, evictionType: evictionType}
	// TAIL will be most recently added
	// HEAD will be earliest added, first to remove
	if l.historyLength == 0 {
		// create linked list
		l.historyHead = historyNode
		l.historyTail = historyNode
		l.historyLength = 1
	} else if l.historyLength == l.maxSize {
		// FIFO head/tail
		prevHistHead := l.historyHead
		nextHistHead := prevHistHead.next
		nextHistHead.prev = nil
		l.historyHead = nextHistHead
		prevHistHead.next = nil
		delete(l.historyLookup, prevHistHead.key)
		prevHistoryTail := l.historyTail
		prevHistoryTail.next = historyNode
		historyNode.prev = prevHistoryTail
		l.historyTail = historyNode
	} else {
		// grow list, this is the new "tail"
		prevHistoryTail := l.historyTail
		prevHistoryTail.next = historyNode
		historyNode.prev = prevHistoryTail
		l.historyTail = historyNode
		l.historyLength = l.historyLength + 1
	}
	oldHistNode, ok := l.historyLookup[historyNode.key]
	if ok {
		l.removeFromHistory(oldHistNode)
		delete(l.historyLookup, oldHistNode.key)
		l.historyLength = l.historyLength - 1
	}
	l.historyLookup[historyNode.key] = historyNode
}

/*SetValue inserts a new cache entry, evicting one if necessary*/
func (l *Lecar) SetValue(k string, v Entry) error {
	lookupNode := &lecarLookupNode{key: k, entry: v}
	lruNode := &lecarLruNode{entryNode: lookupNode}
	lfuNode := &lecarLfuNode{entryNode: lookupNode, accessCount: 1}
	lookupNode.lruNode = lruNode
	lookupNode.lfuNode = lfuNode
	if l.length == 0 {
		// create list head/tail
		l.lruHead = lruNode
		l.lruTail = lruNode
		l.lfuHead = lfuNode
		l.lfuTail = lfuNode
		l.lookup[k] = lookupNode
		l.length = 1
		return nil
	} else if l.length == l.maxSize {
		// evict one entry
		sampleVal := rand.Float64()
		if sampleVal <= l.weightLru {
			// evict by LRU
			prevLruHead := l.lruHead
			evictEntryNode := prevLruHead.entryNode
			delete(l.lookup, evictEntryNode.key)
			l.putInHistory(evictEntryNode, "LRU")
			newLruHead := prevLruHead.next
			newLruHead.prev = nil
			l.lruHead = newLruHead
			// add new value to LRU list
			l.appendToLru(lruNode)
			// remove evicted from LFU list
			l.removeFromLfu(evictEntryNode.lfuNode)
			// insert new value into LFU list
			l.appendToLfu(lfuNode)
		} else {
			// evict by LFU
			prevLfuHead := l.lfuHead
			evictEntryNode := prevLfuHead.entryNode
			delete(l.lookup, evictEntryNode.key)
			l.putInHistory(evictEntryNode, "LFU")
			newLfuHead := prevLfuHead.next
			newLfuHead.prev = nil
			l.lfuHead = newLfuHead
			// add new value to LFU list
			l.appendToLfu(lfuNode)
			// remove evicted from LRU list
			l.removeFromLru(evictEntryNode.lruNode)
			// insert new value into LRU list
			l.appendToLru(lruNode)
		}
		l.lookup[k] = lookupNode

		// length does not change
		return nil
	}
	// grow the LRU list
	l.appendToLru(lruNode)
	// grow the LFU list
	l.appendToLfu(lfuNode)
	// manage lookup
	l.lookup[k] = lookupNode
	l.length = l.length + 1
	return nil
}

func newLecar(size int) *Lecar {
	lk := make(map[string]*lecarLookupNode)
	hk := make(map[string]*lecarHistoryNode)
	return &Lecar{
		maxSize:       size,
		length:        0,
		lruHead:       nil,
		lruTail:       nil,
		lfuHead:       nil,
		lfuTail:       nil,
		lookup:        lk,
		weightLru:     0.5,
		weightLfu:     0.5,
		historyHead:   nil,
		historyTail:   nil,
		historyLookup: hk,
		historyLength: 0,
		lambda:        0.45,
		discount:      0.99,
	}
}
