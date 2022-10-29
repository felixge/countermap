package countermap

import (
	"sync"

	"sync/atomic"
)

func NewSyncMapCounterMap() *SyncMapCounterMap {
	return &SyncMapCounterMap{}
}

type SyncMapCounterMap struct {
	counts sync.Map
}

func (cm *SyncMapCounterMap) Inc(key string) {
	val, ok := cm.counts.Load(key)
	if !ok {
		val, _ = cm.counts.LoadOrStore(key, &atomic.Int64{})
	}
	val.(*atomic.Int64).Add(1)
}

func (cm *SyncMapCounterMap) GetAndReset() map[string]int64 {
	ret := map[string]int64{}
	cm.counts.Range(func(key, val any) bool {
		ret[key.(string)] = val.(*atomic.Int64).Load()
		cm.counts.Delete(key)
		return true
	})
	return ret
}
