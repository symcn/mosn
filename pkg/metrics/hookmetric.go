package metrics

import (
	"sync"

	gometrics "github.com/rcrowley/go-metrics"
	"mosn.io/mosn/pkg/metrics/shm"
)

type hookCounter struct {
	once     sync.Once
	key      string
	prefix   string
	counter  gometrics.Counter
	registry gometrics.Registry
}

func (hc *hookCounter) initial() {
	hc.once.Do(func() {
		hc.counter = hc.registry.GetOrRegister(hc.key, shm.NewShmCounterFunc(hc.prefix+hc.key)).(gometrics.Counter)
	})
}

func (hc *hookCounter) Clear() {
	hc.initial()
	hc.counter.Clear()
}

func (hc *hookCounter) Count() int64 {
	hc.initial()
	return hc.counter.Count()
}

func (hc *hookCounter) Dec(value int64) {
	hc.initial()
	hc.counter.Dec(value)
}

func (hc *hookCounter) Inc(value int64) {
	hc.initial()
	hc.counter.Inc(value)
}

func (hc *hookCounter) Snapshot() gometrics.Counter {
	hc.initial()
	return hc.counter.Snapshot()
}

type hookGauge struct {
	once     sync.Once
	prefix   string
	key      string
	gauge    gometrics.Gauge
	registry gometrics.Registry
}

func (hg *hookGauge) initial() {
	hg.once.Do(func() {
		hg.gauge = hg.registry.GetOrRegister(hg.key, shm.NewShmGaugeFunc(hg.prefix+hg.key)).(gometrics.Gauge)
	})
}

func (hg *hookGauge) Snapshot() gometrics.Gauge {
	hg.initial()
	return hg.gauge.Snapshot()
}

func (hg *hookGauge) Update(value int64) {
	hg.initial()
	hg.gauge.Update(value)
}

func (hg *hookGauge) Value() int64 {
	hg.initial()
	return hg.gauge.Value()
}

type hookHistogram struct {
	once      sync.Once
	key       string
	histogram gometrics.Histogram
	registry  gometrics.Registry
}

func (hh *hookHistogram) initial() {
	hh.once.Do(func() {
		hh.histogram = hh.registry.GetOrRegister(hh.key, func() gometrics.Histogram { return gometrics.NewHistogram(gometrics.NewUniformSample(100)) }).(gometrics.Histogram)
	})
}

func (hh *hookHistogram) Clear() {
	hh.initial()
	hh.histogram.Clear()
}

func (hh *hookHistogram) Count() int64 {
	hh.initial()
	return hh.histogram.Count()
}

func (hh *hookHistogram) Max() int64 {
	hh.initial()
	return hh.histogram.Max()
}

func (hh *hookHistogram) Mean() float64 {
	hh.initial()
	return hh.histogram.Mean()
}

func (hh *hookHistogram) Min() int64 {
	hh.initial()
	return hh.histogram.Min()
}

func (hh *hookHistogram) Percentile(f float64) float64 {
	hh.initial()
	return hh.histogram.Percentile(f)
}

func (hh *hookHistogram) Percentiles(f []float64) []float64 {
	hh.initial()
	return hh.histogram.Percentiles(f)
}

func (hh *hookHistogram) Sample() gometrics.Sample {
	hh.initial()
	return hh.histogram.Sample()
}

func (hh *hookHistogram) Snapshot() gometrics.Histogram {
	hh.initial()
	return hh.histogram.Snapshot()
}

func (hh *hookHistogram) StdDev() float64 {
	hh.initial()
	return hh.histogram.StdDev()
}
func (hh *hookHistogram) Sum() int64 {
	hh.initial()
	return hh.histogram.Sum()
}

func (hh *hookHistogram) Update(value int64) {
	hh.initial()
	hh.histogram.Update(value)
}

func (hh *hookHistogram) Variance() float64 {
	hh.initial()
	return hh.histogram.Variance()
}
