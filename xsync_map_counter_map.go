package countermap

import (
	xsync "github.com/puzpuzpuz/xsync/v2"
)

func NewXSyncMapCounterMap() *XSyncMapCounterMap {
	return &XSyncMapCounterMap{counts: xsync.NewMap()}
}

type XSyncMapCounterMap struct {
	counts *xsync.Map
}

func (cm *XSyncMapCounterMap) Inc(key string) {
	val, ok := cm.counts.Load(key)
	if !ok {
		val, _ = cm.counts.LoadOrStore(key, xsync.NewCounter())
	}
	val.(*xsync.Counter).Add(1)
}

func (cm *XSyncMapCounterMap) GetAndReset() map[string]int64 {
	ret := map[string]int64{}
	cm.counts.Range(func(key string, val any) bool {
		ret[key] = val.(*xsync.Counter).Value()
		cm.counts.Delete(key)
		return true
	})
	return ret
}
