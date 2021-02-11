package models

import (
	"time"

	"github.com/wufe/polo/pkg/utils"
)

type Metrics struct {
	slice *utils.ThreadSafeSlice
}

func NewMetrics() *Metrics {
	return &Metrics{
		slice: &utils.ThreadSafeSlice{
			Elements: []interface{}{},
		},
	}
}

func (m *Metrics) Push(object string, duration time.Duration) {
	m.slice.Push(&Metric{
		Object:   object,
		Duration: duration,
	})
}

func (m *Metrics) ToSlice() []*Metric {
	ret := []*Metric{}
	for _, metric := range m.slice.ToSlice() {
		ret = append(ret, metric.(*Metric))
	}
	return ret
}

type Metric struct {
	Object   string
	Duration time.Duration
}

func NewMetricsForSession(session *Session) func(string) func() {
	return func(object string) func() {
		start := time.Now()
		stopped := false
		return func() {
			if !stopped {
				end := time.Since(start)
				session.Lock()
				defer session.Unlock()
				session.Metrics = append(session.Metrics, &Metric{Object: object, Duration: end})
				stopped = true
			}

		}
	}
}
