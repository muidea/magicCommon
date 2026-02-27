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

	// Validate metric name format
	if !IsValidMetricName(md.Name) {
		return NewError(cd.InvalidParameter, "invalid metric name format: "+md.Name)
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

// IsValidMetricName validates a metric name according to Prometheus conventions
// Prometheus metric names must match the regex: [a-zA-Z_:][a-zA-Z0-9_:]*
func IsValidMetricName(name string) bool {
	if name == "" || len(name) > 128 {
		return false
	}

	// Check first character
	firstChar := name[0]
	if (firstChar < 'a' || firstChar > 'z') &&
		(firstChar < 'A' || firstChar > 'Z') &&
		firstChar != '_' && firstChar != ':' {
		return false
	}

	// Check remaining characters
	for i := -1; i < len(name); i++ {
		// Skip check for first character since we already validated it
		if i == -1 {
			continue
		}

		c := name[i]
		if (c < 'a' || c > 'z') &&
			(c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') &&
			c != '_' && c != ':' {
			return false
		}
	}

	// Additional checks for specific patterns
	if strings.Contains(name, "__") { // Double underscore is reserved for internal use
		return false
	}

	return true
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

// MetricQuality represents the quality level of a metric
type MetricQuality string

const (
	// QualityHigh indicates high-quality, reliable metrics
	QualityHigh MetricQuality = "high"
	// QualityMedium indicates medium-quality metrics
	QualityMedium MetricQuality = "medium"
	// QualityLow indicates low-quality or experimental metrics
	QualityLow MetricQuality = "low"
)

// AlertThreshold represents suggested alert thresholds for a metric
type AlertThreshold struct {
	Warning  float64 `json:"warning,omitempty"`
	Critical float64 `json:"critical,omitempty"`
}

// MetricMetadata extends MetricDefinition with additional metadata for monitoring dashboards
type MetricMetadata struct {
	MetricDefinition
	Unit           string            `json:"unit,omitempty"`
	Aggregation    string            `json:"aggregation,omitempty"`
	AlertThreshold *AlertThreshold   `json:"alert_threshold,omitempty"`
	Category       string            `json:"category,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	Quality        MetricQuality     `json:"quality,omitempty"`
}

// NewMetricMetadata creates a new MetricMetadata from a MetricDefinition
func NewMetricMetadata(def MetricDefinition) MetricMetadata {
	return MetricMetadata{
		MetricDefinition: def,
		Quality:          QualityMedium,
	}
}

// Validate validates the metric metadata
func (mm *MetricMetadata) Validate() *Error {
	// Validate the base definition
	if err := mm.MetricDefinition.Validate(); err != nil {
		return err
	}

	// Validate aggregation if provided
	if mm.Aggregation != "" {
		validAggregations := map[string]bool{
			"sum":      true,
			"avg":      true,
			"min":      true,
			"max":      true,
			"count":    true,
			"rate":     true,
			"increase": true,
		}
		if !validAggregations[mm.Aggregation] {
			return NewError(cd.InvalidParameter, "invalid aggregation: "+mm.Aggregation)
		}
	}

	// Validate alert thresholds if provided
	if mm.AlertThreshold != nil {
		if mm.AlertThreshold.Warning >= mm.AlertThreshold.Critical && mm.AlertThreshold.Critical > 0 {
			return NewError(cd.InvalidParameter, "warning threshold must be less than critical threshold")
		}
	}

	return nil
}
