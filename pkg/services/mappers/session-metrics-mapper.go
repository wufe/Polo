package mappers

import (
	"github.com/wufe/polo/pkg/models"
	"github.com/wufe/polo/pkg/services/output"
)

func MapMetric(model models.Metric) output.Metric {
	return output.Metric{
		Object:   model.Object,
		Duration: int(model.Duration),
	}
}

func MapMetrics(models []models.Metric) []output.Metric {
	ret := []output.Metric{}
	for _, met := range models {
		ret = append(ret, output.Metric{
			Object:   met.Object,
			Duration: int(met.Duration),
		})
	}
	return ret
}
