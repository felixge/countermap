package countermap

import "sync/atomic"

func NewSyncValueCounterMap() *SyncValueCounterMap {
	counts := map[string]*atomic.Int64{}
	cm := &SyncValueCounterMap{}
	cm.counts.Store(&counts)
	return cm
}

type SyncValueCounterMap struct {
	counts atomic.Value
}

func (cm *SyncValueCounterMap) Inc(key string) {
	for {
		oldCounts := cm.counts.Load().(*map[string]*atomic.Int64)
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

func (cm *SyncValueCounterMap) GetAndReset() map[string]int64 {
	for {
		oldCounts := cm.counts.Load().(*map[string]*atomic.Int64)
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
