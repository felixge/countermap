package countermap

import "sync/atomic"

func NewSyncPointerCounterMap() *SyncPointerCounterMap {
	counts := map[string]*atomic.Int64{}
	cm := &SyncPointerCounterMap{}
	cm.counts.Store(&counts)
	return cm
}

type SyncPointerCounterMap struct {
	counts atomic.Pointer[map[string]*atomic.Int64]
}

func (cm *SyncPointerCounterMap) Inc(key string) {
	for {
		oldCounts := cm.counts.Load()
		val, ok := (*oldCounts)[key]
		if ok {
			val.Add(1)
			return
		}

		newCounts := make(map[string]*atomic.Int64)
		for k, v := range *oldCounts {
			newCounts[k] = v
		}
		val = &atomic.Int64{}
		val.Add(1)
		newCounts[key] = val
		if cm.counts.CompareAndSwap(oldCounts, &newCounts) {
			return
		}
	}
}

func (cm *SyncPointerCounterMap) GetAndReset() map[string]int64 {
	for {
		oldCounts := cm.counts.Load()
		retCounts := make(map[string]int64, len(*oldCounts))
		for k, v := range *oldCounts {
			retCounts[k] = v.Load()
		}

		newCounts := map[string]*atomic.Int64{}
		if cm.counts.CompareAndSwap(oldCounts, &newCounts) {
			return retCounts
		}
	}
}
