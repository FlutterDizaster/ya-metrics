package handlers

type AddMetricStorage interface {
	AddMetricValue(kind string, name string, value string) error
}

type GetMetricStorage interface {
	GetMetricValue(kind string, name string) (string, error)
}

type GetAllMetricsStorage interface {
	GetAllMetrics() []struct {
		Name  string
		Kind  string
		Value string
	}
}
