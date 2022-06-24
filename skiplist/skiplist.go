package skiplist

import (
	"fmt"
	"math/bits"
	"math/rand"
	"sync"
	"sync/atomic"
)

const maxHeight = 128

type List[K, V any] struct {
	seed uint64
	cmp  func(a, b K) int

	headVec [maxHeight]Pointer[node[K, V]]
}

func New[K, V any](cmp func(a, b K) int) List[K, V] {
	return List[K, V]{
		cmp: cmp,
	}
}

func (l *List[K, V]) Get(k K) (v V, exists bool) {
	if n := l.getHead(0); n != nil && l.cmp(n.key, k) == 0 {
		return n.value, true
	}
	var vec [maxHeight]*node[K, V]
	l.initPreds(vec[:len(l.headVec)], k)
	l.findPreds(vec[:len(l.headVec)], k)
	if n := vec[0]; n != nil {
		if suc := n.getSuc(0); suc != nil {
			if l.cmp(suc.key, k) == 0 {
				return suc.getValue(), true
			}
		}
	}
	return v, false
}

func (l *List[K, V]) Put(k K, v V) {
	var preds [maxHeight]*node[K, V]
	l.initPreds(preds[:l.height()], k)
	l.findPreds(preds[:l.height()], k)
	h := l.determineHeight(k)
	x := newNode[K, V](h, k)
	x.setValue(v)

	get := func(i int) *node[K, V] {
		if preds[i] == nil {
			return l.getHead(i)
		} else {
			return preds[i].getSuc(i)
		}
	}
	cas := func(i int, prev, next *node[K, V]) *node[K, V] {
		if preds[i] == nil {
			return l.casHead(i, prev, next)
		} else {
			return preds[i].casSuc(i, prev, next)
		}
	}
	for i := 0; i < h; i++ {
		suc := get(i)
		//log.Printf("inserting level %d, suc=%v x=%v", i, suc, x)
		for suc != x {
			x.setSuc(i, suc)
			// if there is no successor, then don't compare
			if suc == nil {
				suc = cas(i, suc, x)
				continue
			}
			c := l.cmp(suc.key, x.key)
			switch {
			case c == 0:
				// successor is x
				// the node already exists, update the value (the key never changes)
				suc.setValue(x.value)
				return
			case c > 0:
				// successor is still after x, do the swap.
				suc = cas(i, suc, x)
			case c < 0:
				// successor is before x
				// someone else got to this spot first, and inserted a node > pred
				// need to move forward
				preds[i] = suc
				suc = suc.getSuc(i)
			}
		}
	}
}

func (l *List[K, V]) Delete(k K) {
	var preds [maxHeight]*node[K, V]
	h := l.height()
	l.initPreds(preds[:h], k)
	l.findPreds(preds[:h], k)
	get := func(i int) *node[K, V] {
		if preds[i] == nil {
			return l.getHead(i)
		} else {
			return preds[i].getSuc(i)
		}
	}
	cas := func(i int, prev, next *node[K, V]) *node[K, V] {
		if preds[i] == nil {
			return l.casHead(i, prev, next)
		} else {
			return preds[i].casSuc(i, prev, next)
		}
	}
	for i := 0; i < h; i++ {
		for {
			x := get(i)
			if x == nil {
				break // success: nothing there
			}
			c := l.cmp(x.key, k)
			if c < 0 {
				// need to keep traversing
				preds[i] = x
				continue
			} else if c > 0 {
				break // success: already gone
			} else {
				// x is a match, time to delete
				suc := x.getSuc(i)
				predSuc := cas(i, x, suc)
				if predSuc != x {
					// if the predecessor's successor is not x, then x is gone.
					break
				}
			}
		}
	}
}

func (l *List[K, V]) ForEach(gteq *K, fn func(k K, v V) bool) {
	l.forEach(gteq, func(n *node[K, V]) bool {
		return fn(n.key, n.getValue())
	})
}

func (l *List[K, V]) forEach(gteq *K, fn func(n *node[K, V]) bool) {
	var preds [maxHeight]*node[K, V]
	if gteq != nil {
		h := l.height()
		l.initPreds(preds[:h], *gteq)
		l.findPreds(preds[:h], *gteq)
	} else {
		preds[0] = l.getHead(0)
	}
	for n := preds[0]; n != nil; n = n.getSuc(0) {
		if !fn(n) {
			return
		}
	}
}

func (l *List[K, V]) getHead(i int) *node[K, V] {
	return l.headVec[i].Load()
}

func (l *List[K, V]) casHead(i int, prev, next *node[K, V]) (ret *node[K, V]) {
	if l.headVec[i].CompareAndSwap(prev, next) {
		return next
	}
	return l.headVec[i].Load()
}

func (l *List[K, V]) height() int {
	for i := 1; i < len(l.headVec); i++ {
		if l.getHead(i) == nil {
			return i
		}
	}
	return len(l.headVec)
}

// initPreds fills vec with values from headVec < k
func (l *List[K, V]) initPreds(vec []*node[K, V], k K) {
	// fill vec with nodes < k
	for i := range vec {
		if n := l.getHead(i); n != nil && l.cmp(n.key, k) < 0 {
			vec[i] = n
		}
	}
}

// findPreds fills vec with the immediate predecessor on every level
func (l *List[K, V]) findPreds(preds []*node[K, V], k K) {
	// l.assertPreds(preds, k)
	// improve vec until it contains the immediate predecessors at each level
	for i := len(preds) - 1; i >= 0; i-- {
		// attempt to improve vec[i]
		// first follow the successors, until they are bigger, or nil
		for preds[i] != nil {
			if suc := preds[i].getSuc(i); suc != nil {
				if c := l.cmp(suc.key, k); c < 0 {
					preds[i] = suc
					continue
				}
			}
			break
		}
		// then improve the lower predecessors
		if preds[i] != nil {
			for j := 0; j < i; j++ {
				if suc := preds[i].getSuc(j); suc != nil {
					if c := l.cmp(suc.key, k); c < 0 {
						preds[j] = suc
					}
				}
			}
		}
	}
	// l.assertPreds(preds, k)
}

func (l *List[K, V]) assertPreds(preds []*node[K, V], k K) {
	for i := range preds {
		if preds[i] == nil {
			continue
		}
		if l.cmp(preds[i].key, k) >= 0 {
			panic(fmt.Sprintf("pred=%v for k=%v", preds[i].key, k))
		}
	}
}

func (l *List[K, V]) determineHeight(k K) int {
	seed := atomic.AddUint64(&l.seed, 1)
	rng := rand.New(rand.NewSource(int64(seed)))
	return determineHeight(rng)
}

type node[K, V any] struct {
	sucs []Pointer[node[K, V]]

	key   K
	mu    sync.RWMutex
	value V
}

func newNode[K, V any](height int, k K) *node[K, V] {
	return &node[K, V]{
		sucs: make([]Pointer[node[K, V]], height),
		key:  k,
	}
}

func (n *node[K, V]) getSuc(level int) *node[K, V] {
	return n.sucs[level].Load()
}

func (n *node[K, V]) casSuc(level int, prev, next *node[K, V]) *node[K, V] {
	if n.sucs[level].CompareAndSwap(prev, next) {
		return next
	}
	return n.sucs[level].Load()
}

func (n *node[K, V]) setSuc(level int, x *node[K, V]) {
	n.sucs[level].Store(x)
}

func (n *node[K, V]) setValue(v V) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.value = v
}

func (n *node[K, V]) getValue() V {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.value
}

func (n *node[K, V]) String() string {
	return fmt.Sprintf("{key: %v, h: %d, next: %v }", n.key, len(n.sucs), n.getSuc(0))
}

func determineHeight(rng *rand.Rand) (ret int) {
	for {
		x := rng.Uint64()
		lz := bits.LeadingZeros64(x)
		ret += lz
		if lz < 64 {
			break
		}
	}
	ret++
	if ret > maxHeight {
		ret = maxHeight
	}
	return ret
}
