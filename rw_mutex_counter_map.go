package countermap

import (
	"sync"
	"sync/atomic"
)

func NewRWMutexCounterMap() *RWMutexCounterMap {
	return &RWMutexCounterMap{counts: map[string]*atomic.Int64{}}
}

type RWMutexCounterMap struct {
	lock   sync.RWMutex
	counts map[string]*atomic.Int64
}

func (cm *RWMutexCounterMap) Inc(key string) {
	cm.lock.RLock()
	val, ok := cm.counts[key]
	cm.lock.RUnlock()

	if !ok {
		cm.lock.Lock()
		defer cm.lock.Unlock()
		val = &atomic.Int64{}
		cm.counts[key] = val
	}

	val.Add(1)
}

func (cm *RWMutexCounterMap) GetAndReset() map[string]int64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	counts := make(map[string]int64, len(cm.counts))
	for k, v := range cm.counts {
		counts[k] = v.Load()
	}
	cm.counts = make(map[string]*atomic.Int64)
	return counts
}
