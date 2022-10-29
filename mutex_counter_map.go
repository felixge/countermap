package countermap

import "sync"

func NewMutexCounterMap() *MutexCounterMap {
	return &MutexCounterMap{counts: map[string]int64{}}
}

type MutexCounterMap struct {
	lock   sync.Mutex
	counts map[string]int64
}

func (cm *MutexCounterMap) Inc(key string) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.counts[key]++
}

func (cm *MutexCounterMap) GetAndReset() map[string]int64 {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	ret := cm.counts
	cm.counts = map[string]int64{}
	return ret
}
