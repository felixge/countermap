package countermap

func NewUnguardedCounterMap() *UnguardedCounterMap {
	return &UnguardedCounterMap{counts: map[string]int64{}}
}

type UnguardedCounterMap struct {
	counts map[string]int64
}

func (cm *UnguardedCounterMap) Inc(key string) {
	cm.counts[key]++
}

func (cm *UnguardedCounterMap) GetAndReset() map[string]int64 {
	ret := cm.counts
	cm.counts = map[string]int64{}
	return ret
}
