package types

import (
	"strings"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

// MetricType represents the type of metric
type MetricType string

const (
	// CounterMetric represents a cumulative metric that only increases
	CounterMetric MetricType = "counter"
	// GaugeMetric represents a metric that can go up and down
	GaugeMetric MetricType = "gauge"
	// HistogramMetric represents a metric that samples observations
	HistogramMetric MetricType = "histogram"
	// SummaryMetric represents a metric that calculates quantiles
	SummaryMetric MetricType = "summary"
)

// Metric represents a single metric with labels and value
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
	Timestamp   time.Time         `json:"timestamp"`
	Description string            `json:"description,omitempty"`
}

// MetricDefinition defines a metric's structure and behavior
type MetricDefinition struct {
	Name        string              `json:"name"`
	Type        MetricType          `json:"type"`
	Help        string              `json:"help"`
	LabelNames  []string            `json:"label_names"`
	Buckets     []float64           `json:"buckets,omitempty"`    // For histograms
	Objectives  map[float64]float64 `json:"objectives,omitempty"` // For summaries
	MaxAge      time.Duration       `json:"max_age,omitempty"`
	ConstLabels map[string]string   `json:"const_labels,omitempty"`
}

// NewMetric creates a new metric with the current timestamp
func NewMetric(name string, metricType MetricType, value float64, labels map[string]string) Metric {
	return Metric{
		Name:      name,
		Type:      metricType,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}
}

// NewCounter creates a new counter metric
func NewCounter(name string, value float64, labels map[string]string) Metric {
	return NewMetric(name, CounterMetric, value, labels)
}

// NewGauge creates a new gauge metric
func NewGauge(name string, value float64, labels map[string]string) Metric {
	return NewMetric(name, GaugeMetric, value, labels)
}

// NewCounterDefinition creates a new counter metric definition
func NewCounterDefinition(name, help string, labelNames []string, constLabels map[string]string) MetricDefinition {
	return MetricDefinition{
		Name:        name,
		Type:        CounterMetric,
		Help:        help,
		LabelNames:  labelNames,
		ConstLabels: constLabels,
	}
}

// NewGaugeDefinition creates a new gauge metric definition
func NewGaugeDefinition(name, help string, labelNames []string, constLabels map[string]string) MetricDefinition {
	return MetricDefinition{
		Name:        name,
		Type:        GaugeMetric,
		Help:        help,
		LabelNames:  labelNames,
		ConstLabels: constLabels,
	}
}

// NewHistogramDefinition creates a new histogram metric definition
func NewHistogramDefinition(name, help string, labelNames []string, buckets []float64, constLabels map[string]string) MetricDefinition {
	return MetricDefinition{
		Name:        name,
		Type:        HistogramMetric,
		Help:        help,
		LabelNames:  labelNames,
		Buckets:     buckets,
		ConstLabels: constLabels,
	}
}

// NewSummaryDefinition creates a new summary metric definition
func NewSummaryDefinition(name, help string, labelNames []string, objectives map[float64]float64, maxAge time.Duration, constLabels map[string]string) MetricDefinition {
	return MetricDefinition{
		Name:        name,
		Type:        SummaryMetric,
		Help:        help,
		LabelNames:  labelNames,
		Objectives:  objectives,
		MaxAge:      maxAge,
		ConstLabels: constLabels,
	}
}

// Validate validates the metric definition
func (md *MetricDefinition) Validate() *Error {
	if md.Name == "" {
		return NewError(cd.InvalidParameter, "metric name cannot be empty")
	}

	if md.Help == "" {
		return NewError(cd.InvalidParameter, "metric help text cannot be empty")
	}

	// Validate label names
	for _, label := range md.LabelNames {
		if label == "" {
			return NewError(cd.InvalidParameter, "label name cannot be empty")
		}
	}

	// Validate type-specific constraints
	switch md.Type {
	case HistogramMetric:
		if len(md.Buckets) == 0 {
			return NewError(cd.InvalidParameter, "histogram must have at least one bucket")
		}
		// Ensure buckets are sorted
		for i := 1; i < len(md.Buckets); i++ {
			if md.Buckets[i] <= md.Buckets[i-1] {
				return NewError(cd.InvalidParameter, "histogram buckets must be in increasing order")
			}
		}
	case SummaryMetric:
		if len(md.Objectives) == 0 {
			return NewError(cd.InvalidParameter, "summary must have at least one objective")
		}
		for quantile := range md.Objectives {
			if quantile < 0 || quantile > 1 {
				return NewError(cd.InvalidParameter, "summary quantile must be between 0 and 1")
			}
		}
		if md.MaxAge <= 0 {
			return NewError(cd.InvalidParameter, "summary max age must be positive")
		}
	}

	return nil
}

// GetFullName returns the full metric name with namespace prefix if provided
func (md *MetricDefinition) GetFullName(namespace string) string {
	if namespace == "" {
		return md.Name
	}

	// Check if name already starts with namespace
	if strings.HasPrefix(md.Name, namespace+"_") {
		return md.Name
	}

	return namespace + "_" + md.Name
}
