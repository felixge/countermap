package countermap

import (
	xsync "github.com/puzpuzpuz/xsync/v2"
)

func NewXSyncMapCounterMap() *XSyncMapCounterMap {
	return &XSyncMapCounterMap{counts: xsync.NewMapOf[*xsync.Counter]()}
}

type XSyncMapCounterMap struct {
	counts *xsync.MapOf[string, *xsync.Counter]
}

func (cm *XSyncMapCounterMap) Inc(key string) {
	c, _ := cm.counts.LoadOrCompute(key, func() *xsync.Counter {
		return xsync.NewCounter()
	})
	c.Inc()
}

func (cm *XSyncMapCounterMap) GetAndReset() map[string]int64 {
	ret := make(map[string]int64, cm.counts.Size())
	cm.counts.Range(func(key string, val *xsync.Counter) bool {
		ret[key] = val.Value()
		return true
	})
	cm.counts.Clear()
	return ret
}
