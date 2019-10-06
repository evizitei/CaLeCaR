package cache

import (
	"errors"
	"math"
	"math/rand"
)

type calecarLruNode struct {
	prev      *calecarLruNode
	next      *calecarLruNode
	entryNode *calecarLookupNode
}

type calecarLfuNode struct {
	prev        *calecarLfuNode
	next        *calecarLfuNode
	accessCount int
	entryNode   *calecarLookupNode
}

type calecarLcrNode struct {
	prev      *calecarLcrNode
	next      *calecarLcrNode
	entryNode *calecarLookupNode
}

type calecarLookupNode struct {
	key     string
	entry   Entry
	lruNode *calecarLruNode
	lfuNode *calecarLfuNode
	lcrNode *calecarLcrNode
}

type calecarHistoryNode struct {
	key          string
	evictionType string
	next         *calecarHistoryNode
	prev         *calecarHistoryNode
}

/*Calecar balances a frequency and recency distribution
https://www.usenix.org/system/files/conference/hotstorage18/hotstorage18-paper-vietri.pdf
*/
type Calecar struct {
	maxSize       int
	length        int
	lruHead       *calecarLruNode
	lruTail       *calecarLruNode
	lfuHead       *calecarLfuNode
	lfuTail       *calecarLfuNode
	lcrHead       *calecarLcrNode
	lcrTail       *calecarLcrNode
	weightLru     float64
	weightLfu     float64
	weightLcr     float64
	lookup        map[string]*calecarLookupNode
	debug         bool
	historyLookup map[string]*calecarHistoryNode
	historyLength int
	historyHead   *calecarHistoryNode
	historyTail   *calecarHistoryNode
	lambda        float64
	discount      float64
}

func (c *Calecar) updateAlgoWeights(node *calecarHistoryNode) {
	regret := 1.0
	histPosition := 1
	pointerNode := node
	for {
		if pointerNode == c.historyTail {
			// discount is adjusted correctly now
			break
		}
		pointerNode = pointerNode.next
		regret = regret * c.discount
		histPosition = histPosition + 1
	}
	// this will adjust a given weight *DOWN* by an amount inversely
	// related to length of time in history
	adjustCoefficient := (1 / math.Pow(math.E, (c.lambda*regret)))
	wLru := c.weightLru
	wLfu := c.weightLfu
	wLcr := c.weightLcr
	if node.evictionType == "LRU" {
		wLru = wLru * adjustCoefficient
	} else if node.evictionType == "LFU" {
		wLfu = wLfu * adjustCoefficient
	} else if node.evictionType == "LCR" {
		wLcr = wLcr * adjustCoefficient
	}
	normConst := (wLfu + wLru + wLcr)
	c.weightLfu = wLfu / normConst
	c.weightLru = wLru / normConst
	c.weightLcr = wLcr / normConst
}

/*KeyPresent is true if the key is in the cache right now*/
func (c *Calecar) KeyPresent(k string) bool {
	_, ok := c.lookup[k]
	if !ok {
		// key not in cache, check history and penalize if present
		historyNode, hOk := c.historyLookup[k]
		if hOk {
			// in history, penalize an algorithm
			c.updateAlgoWeights(historyNode)
		}
	}
	return ok
}

/*GetValue will return the entry if present in the lookup*/
func (c *Calecar) GetValue(k string) (Entry, error) {
	lookupNode, ok := c.lookup[k]
	if !ok {
		return Entry{}, errors.New("Key not present in lookup hash")
	}
	lruNode := lookupNode.lruNode
	// LRU: promote entry to most recently accessed
	if lruNode == c.lruTail {
		// do nothing, it's already most recently accessed
	} else if lruNode == c.lruHead {
		// just move head to tail
		newHead := lruNode.next
		newHead.prev = nil
		c.lruHead = newHead
		prevTail := c.lruTail
		lruNode.prev = prevTail
		prevTail.next = lruNode
		c.lruTail = lruNode
		lruNode.next = nil
	} else {
		// in the middle, stitch two nodes together and move to tail
		oldPrev := lruNode.prev
		oldNext := lruNode.next
		oldPrev.next = oldNext
		oldNext.prev = oldPrev
		prevTail := c.lruTail
		lruNode.prev = prevTail
		lruNode.next = nil
		prevTail.next = lruNode
		c.lruTail = lruNode
	}
	// LFU: increment access count and reorder
	lfuNode := lookupNode.lfuNode
	lfuNode.accessCount = lfuNode.accessCount + 1
	// move node to the right until it is accessed more
	// than prev and less than next, or until it is the tail
	if lfuNode == c.lfuTail {
		// do nothing, it's already the most frequently accessed
	} else {
		c.reorderLfuList(lfuNode)
	}
	// LCR: Do nothing; access does not change cost
	return lookupNode.entry, nil
}

func (c *Calecar) reorderLfuList(node *calecarLfuNode) {
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
				c.lfuHead = newHead
				if rightHead == nil {
					// node is now tail
					c.lfuTail = node
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
					c.lfuTail = node
					return
				}
				rightHead.prev = node
			}
		} else {
			return
		}
	}
}

func (c *Calecar) reorderLcrList(node *calecarLcrNode) {
	for {
		if node.entryNode.entry.cost >= node.next.entryNode.entry.cost {
			// swap positions
			if node.prev == nil {
				// node is currently HEAD
				newHead := node.next
				rightHead := newHead.next
				newHead.prev = nil
				node.next = rightHead
				node.prev = newHead
				newHead.next = node
				c.lcrHead = newHead
				if rightHead == nil {
					// node is now tail
					c.lcrTail = node
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
					c.lcrTail = node
					return
				}
				rightHead.prev = node
			}
		} else {
			return
		}
	}
}

func (c *Calecar) removeFromLru(node *calecarLruNode) {
	if node == c.lruHead {
		newLruHead := node.next
		newLruHead.prev = nil
		c.lruHead = newLruHead
		node.next = nil
	} else if node == c.lruTail {
		newLruTail := node.prev
		newLruTail.next = nil
		c.lruTail = newLruTail
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

func (c *Calecar) removeFromLfu(node *calecarLfuNode) {
	if node == c.lfuHead {
		newLfuHead := node.next
		newLfuHead.prev = nil
		c.lfuHead = newLfuHead
		node.next = nil
	} else if node == c.lfuTail {
		newLfuTail := node.prev
		newLfuTail.next = nil
		c.lfuTail = newLfuTail
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

func (c *Calecar) removeFromLcr(node *calecarLcrNode) {
	if node == c.lcrHead {
		newLcrHead := node.next
		newLcrHead.prev = nil
		c.lcrHead = newLcrHead
		node.next = nil
	} else if node == c.lcrTail {
		newLcrTail := node.prev
		newLcrTail.next = nil
		c.lcrTail = newLcrTail
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

func (c *Calecar) removeFromHistory(histNode *calecarHistoryNode) {
	if histNode == c.historyHead {
		newHistHead := histNode.next
		newHistHead.prev = nil
		c.historyHead = newHistHead
		histNode.next = nil
	} else if histNode == c.historyTail {
		newHistTail := histNode.prev
		newHistTail.next = nil
		c.historyTail = newHistTail
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

func (c *Calecar) appendToLfu(lfuNode *calecarLfuNode) {
	oldLfuHead := c.lfuHead
	lfuNode.next = oldLfuHead
	oldLfuHead.prev = lfuNode
	c.lfuHead = lfuNode
	c.reorderLfuList(lfuNode)
}

func (c *Calecar) appendToLcr(lcrNode *calecarLcrNode) {
	oldLcrHead := c.lcrHead
	lcrNode.next = oldLcrHead
	oldLcrHead.prev = lcrNode
	c.lcrHead = lcrNode
	c.reorderLcrList(lcrNode)
}

func (c *Calecar) appendToLru(lruNode *calecarLruNode) {
	prevLruTail := c.lruTail
	prevLruTail.next = lruNode
	lruNode.prev = prevLruTail
	c.lruTail = lruNode
}

func (c *Calecar) putInHistory(entryNode *calecarLookupNode, evictionType string) {
	historyNode := &calecarHistoryNode{key: entryNode.key, evictionType: evictionType}
	// TAIL will be most recently added
	// HEAD will be earliest added, first to remove
	if c.historyLength == 0 {
		// create linked list
		c.historyHead = historyNode
		c.historyTail = historyNode
	} else if c.historyLength == c.maxSize {
		// FIFO head/tail
		prevHistHead := c.historyHead
		nextHistHead := prevHistHead.next
		nextHistHead.prev = nil
		c.historyHead = nextHistHead
		prevHistHead.next = nil
		delete(c.historyLookup, prevHistHead.key)
		prevHistoryTail := c.historyTail
		prevHistoryTail.next = historyNode
		historyNode.prev = prevHistoryTail
		c.historyTail = historyNode
	} else {
		// grow list, this is the new "tail"
		prevHistoryTail := c.historyTail
		prevHistoryTail.next = historyNode
		historyNode.prev = prevHistoryTail
		c.historyTail = historyNode
	}
	oldHistNode, ok := c.historyLookup[historyNode.key]
	if ok {
		c.removeFromHistory(oldHistNode)
		delete(c.historyLookup, oldHistNode.key)
	}
	c.historyLookup[historyNode.key] = historyNode
}

/*SetValue inserts a new cache entry, evicting one if necessary*/
func (c *Calecar) SetValue(k string, v Entry) error {
	lookupNode := &calecarLookupNode{key: k, entry: v}
	lruNode := &calecarLruNode{entryNode: lookupNode}
	lfuNode := &calecarLfuNode{entryNode: lookupNode, accessCount: 1}
	lcrNode := &calecarLcrNode{entryNode: lookupNode}
	lookupNode.lruNode = lruNode
	lookupNode.lfuNode = lfuNode
	lookupNode.lcrNode = lcrNode
	if c.length == 0 {
		// create list head/tail
		c.lruHead = lruNode
		c.lruTail = lruNode
		c.lfuHead = lfuNode
		c.lfuTail = lfuNode
		c.lcrHead = lcrNode
		c.lcrTail = lcrNode
		c.lookup[k] = lookupNode
		c.length = 1
		return nil
	} else if c.length == c.maxSize {
		// evict one entry
		sampleVal := rand.Float64()
		if sampleVal <= c.weightLru {
			// evict by LRU
			prevLruHead := c.lruHead
			evictEntryNode := prevLruHead.entryNode
			delete(c.lookup, evictEntryNode.key)
			c.putInHistory(evictEntryNode, "LRU")
			newLruHead := prevLruHead.next
			newLruHead.prev = nil
			c.lruHead = newLruHead
			prevLruHead.next = nil
			// add new value to LRU list
			c.appendToLru(lruNode)
			// remove evicted from LFU and LCR lists
			c.removeFromLfu(evictEntryNode.lfuNode)
			c.removeFromLcr(evictEntryNode.lcrNode)
			// insert new value into LFU and LCR lists
			c.appendToLfu(lfuNode)
			c.appendToLcr(lcrNode)
		} else if sampleVal <= (c.weightLru + c.weightLfu) {
			// evict by LFU
			prevLfuHead := c.lfuHead
			evictEntryNode := prevLfuHead.entryNode
			delete(c.lookup, evictEntryNode.key)
			c.putInHistory(evictEntryNode, "LFU")
			newLfuHead := prevLfuHead.next
			newLfuHead.prev = nil
			c.lfuHead = newLfuHead
			prevLfuHead.next = nil
			// add new value to LFU list
			c.appendToLfu(lfuNode)
			// remove evicted from LRU list
			c.removeFromLru(evictEntryNode.lruNode)
			c.removeFromLcr(evictEntryNode.lcrNode)
			// insert new value into LRU list
			c.appendToLru(lruNode)
			c.appendToLcr(lcrNode)
		} else {
			// evict by LCR
			prevLcrHead := c.lcrHead
			evictEntryNode := prevLcrHead.entryNode
			delete(c.lookup, evictEntryNode.key)
			c.putInHistory(evictEntryNode, "LCR")
			newLcrHead := prevLcrHead.next
			newLcrHead.prev = nil
			c.lcrHead = newLcrHead
			prevLcrHead.next = nil
			// add new val to LCR
			c.appendToLcr(lcrNode)
			// remove evicted from LRU and LFU list
			c.removeFromLru(evictEntryNode.lruNode)
			c.removeFromLfu(evictEntryNode.lfuNode)
			// get new val into other lists
			c.appendToLru(lruNode)
			c.appendToLfu(lfuNode)
		}
		c.lookup[k] = lookupNode

		// length does not change
		return nil
	}
	// grow the lists
	c.appendToLru(lruNode)
	c.appendToLfu(lfuNode)
	c.appendToLcr(lcrNode)
	// manage lookup
	c.lookup[k] = lookupNode
	c.length = c.length + 1
	return nil
}

func newCalecar(size int) *Calecar {
	lk := make(map[string]*calecarLookupNode)
	hk := make(map[string]*calecarHistoryNode)
	return &Calecar{
		maxSize:       size,
		length:        0,
		lruHead:       nil,
		lruTail:       nil,
		lfuHead:       nil,
		lfuTail:       nil,
		lcrHead:       nil,
		lcrTail:       nil,
		lookup:        lk,
		weightLru:     0.33,
		weightLfu:     0.33,
		weightLcr:     0.33,
		historyHead:   nil,
		historyTail:   nil,
		historyLookup: hk,
		historyLength: 0,
		lambda:        0.45,
		discount:      0.99,
	}
}
