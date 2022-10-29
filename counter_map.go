package countermap

type CounterMap interface {
	Inc(key string)
	GetAndReset() map[string]int64
}
