package go_deep_copy

import (
	"github.com/modern-go/reflect2"
	"sync"
	"sync/atomic"
	"unsafe"
)

// RCU 依据 Read Copy Update 原理实现
type RCU struct {
	lock sync.Mutex
	m    unsafe.Pointer
}

func NewRCU() (c *RCU) {
	return &RCU{
		lock: sync.Mutex{},
		m: unsafe.Pointer(&linerMap{
			n: 0,
			m: _InitCapacity - 1,
			b: make([]mapEntry, _InitCapacity),
		}),
	}
}

func (c *RCU) Get(key reflect2.Type) (v any, ok bool) {
	m := (*linerMap)(atomic.LoadPointer(&c.m))
	res := m.get(key)
	return res, res != nil
}

func (c *RCU) Set(key reflect2.Type, v any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	m := (*linerMap)(atomic.LoadPointer(&c.m))
	atomic.StorePointer(&c.m, unsafe.Pointer(m.add(key, v)))
}

func (c *RCU) GetOrSet(key reflect2.Type, newV any) (v any, loaded bool) {
	got, ok := c.Get(key)
	if ok {
		return got, true
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	// double check
	m := (*linerMap)(atomic.LoadPointer(&c.m))
	res := m.get(key)
	if res != nil {
		return res, true
	}

	m2 := m.add(key, newV)
	atomic.StorePointer(&c.m, unsafe.Pointer(m2))

	return newV, false
}

/** 线性探测的开放寻址 Map **/

const (
	_LoadFactor   = 0.5
	_InitCapacity = 4096 // must be a power of 2
)

type linerMap struct {
	n uint64 // 实际元素个数
	m uint32 // capacity
	b []mapEntry
}

type mapEntry struct {
	vt reflect2.Type
	fn any
}

func newProgramMap() *linerMap {
	return &linerMap{
		n: 0,
		m: _InitCapacity - 1,
		b: make([]mapEntry, _InitCapacity),
	}
}

func (self *linerMap) copy() *linerMap {
	fork := &linerMap{
		n: self.n,
		m: self.m,
		b: make([]mapEntry, len(self.b)),
	}
	for i, f := range self.b {
		fork.b[i] = f
	}
	return fork
}

func (self *linerMap) get(vt reflect2.Type) any {
	i := self.m + 1
	h := uint32(vt.RType())
	p := h & self.m

	/* linear probing */
	for ; i > 0; i-- {
		if b := self.b[p]; b.vt == vt {
			return b.fn
		} else if b.vt == nil {
			break
		} else {
			p = (p + 1) & self.m
		}
	}

	/* not found */
	return nil
}

func (self *linerMap) add(vt reflect2.Type, fn any) *linerMap {
	p := self.copy()
	f := float64(atomic.LoadUint64(&p.n)+1) / float64(p.m+1)

	/* check for load factor */
	if f > _LoadFactor {
		p = p.rehash()
	}

	/* insert the value */
	p.insert(vt, fn)
	return p
}

func (self *linerMap) rehash() *linerMap {
	c := (self.m + 1) << 1
	r := &linerMap{m: c - 1, b: make([]mapEntry, int(c))}

	/* rehash every entry */
	for i := uint32(0); i <= self.m; i++ {
		if b := self.b[i]; b.vt != nil {
			r.insert(b.vt, b.fn)
		}
	}

	/* rebuild successful */
	return r
}

func (self *linerMap) insert(vt reflect2.Type, fn any) {
	h := uint32(vt.RType())
	p := h & self.m

	/* linear probing */
	for i := uint32(0); i <= self.m; i++ {
		if b := &self.b[p]; b.vt != nil {
			p += 1
			p &= self.m
		} else {
			b.vt = vt
			b.fn = fn
			atomic.AddUint64(&self.n, 1)
			return
		}
	}

	/* should never happens */
	panic("no available slots")
}
