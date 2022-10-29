package countermap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var counterMapImplementations = []struct {
	Name string
	New  func() CounterMap
}{
	//{"Unguarded", func() CounterMap { return NewUnguardedCounterMap() }},
	//{"sync.RWMutex", func() CounterMap { return NewRWMutexCounterMap() }},

	{"sync.Mutex", func() CounterMap { return NewMutexCounterMap() }},
	{"sync.Map", func() CounterMap { return NewSyncMapCounterMap() }},
	{"sync.Value", func() CounterMap { return NewSyncValueCounterMap() }},
	{"sync.Pointer", func() CounterMap { return NewSyncPointerCounterMap() }},
}

func BenchmarkCounterMap(b *testing.B) {
	endpoints := make([]string, 10)
	for i := range endpoints {
		endpoints[i] = fmt.Sprintf("endpoint-%d", i)
	}

	for _, impl := range counterMapImplementations {
		b.Run(impl.Name, func(b *testing.B) {
			b.ReportAllocs()
			cm := impl.New()
			b.RunParallel(func(p *testing.PB) {
				i := 0
				for p.Next() {
					cm.Inc(endpoints[i%len(endpoints)])
					i++
				}
			})

			// The benchmark above is constructed so that endpoints should exhibit
			// monotonically decreasing hit counts. If this invariant is violated
			// the implementation is buggy and we fail the benchmark.
			counts := cm.GetAndReset()
			for i := 1; i < len(endpoints); i++ {
				endpoint := endpoints[i]
				prevEndpoint := endpoints[i-1]
				if counts[endpoint] > counts[prevEndpoint] {
					b.Fatalf("%q: %d > %q:%d", endpoint, counts[endpoint], prevEndpoint, counts[prevEndpoint])
				}
			}
		})
	}
}

func TestCounterMap(t *testing.T) {
	for _, impl := range counterMapImplementations {
		t.Run(impl.Name, func(t *testing.T) {
			cm := impl.New()
			cm.Inc("foo")
			cm.Inc("foo")
			cm.Inc("bar")
			require.Equal(t, map[string]int64{"foo": 2, "bar": 1}, cm.GetAndReset())
			cm.Inc("foobar")
			require.Equal(t, map[string]int64{"foobar": 1}, cm.GetAndReset())
			require.Equal(t, map[string]int64{}, cm.GetAndReset())
		})
	}
}
