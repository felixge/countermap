package countermap

import (
	"encoding/binary"
	"fmt"
	"testing"

	"sync/atomic"

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
	{"xsync.Map", func() CounterMap { return NewXSyncMapCounterMap() }},
	{"sync.Value", func() CounterMap { return NewSyncValueCounterMap() }},
	{"sync.Pointer", func() CounterMap { return NewSyncPointerCounterMap() }},
}

func BenchmarkCounterMap(b *testing.B) {
	b.Run("ExpectedCase", func(b *testing.B) {
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
	})

	b.Run("WorstCase", func(b *testing.B) {
		for _, impl := range counterMapImplementations {
			b.Run(impl.Name, func(b *testing.B) {
				b.ReportAllocs()
				cm := impl.New()
				var id atomic.Int64
				var totalCount atomic.Int64
				b.RunParallel(func(p *testing.PB) {
					key := binary.BigEndian.AppendUint64([]byte(fmt.Sprintf("endpoint-%03d-", id.Add(1))), 0)
					i := uint64(0)
					for p.Next() {
						binary.BigEndian.PutUint64(key[len(key)-8:], i)
						cm.Inc(string(key))
						totalCount.Add(1)
						i++
					}
				})

				var sum int64
				counts := cm.GetAndReset()
				for _, count := range counts {
					require.Equal(b, count, int64(1))
					sum += count
				}
				require.Equal(b, int64(len(counts)), totalCount.Load())
				require.Equal(b, totalCount.Load(), sum)
			})
		}
	})

	b.Run("ConsistentSnapshot", func(b *testing.B) {
		for _, impl := range counterMapImplementations {
			b.Run(impl.Name, func(b *testing.B) {
				b.ReportAllocs()

				cm := impl.New()
				var id atomic.Int64
				var totalCount atomic.Int64

				var fail string
				doneCh := make(chan struct{})
				go func() {
					var sumCounts int64
					for {
						select {
						case <-doneCh:
							return
						default:
							counts := cm.GetAndReset()
							sumCounts += int64(len(counts))
							require.GreaterOrEqual(b, sumCounts, totalCount.Load())
						}
					}
				}()

				b.RunParallel(func(p *testing.PB) {
					key := binary.BigEndian.AppendUint64([]byte(fmt.Sprintf("endpoint-%03d-", id.Add(1))), 0)
					i := uint64(0)
					for p.Next() {
						binary.BigEndian.PutUint64(key[len(key)-8:], i)
						cm.Inc(string(key))
						totalCount.Add(1)
						i++
					}
				})

				close(doneCh)
				require.Empty(b, fail)
			})
		}
	})
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
