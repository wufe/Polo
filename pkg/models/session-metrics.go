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

type Metric struct {
	Object   string
	Duration time.Duration
}
