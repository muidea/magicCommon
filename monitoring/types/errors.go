package types

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
)

// Error is an alias for cd.Error for convenience
type Error = cd.Error

// Monitoring-specific error codes
const (
	// MetricAlreadyRegistered indicates a metric with the same name is already registered
	MetricAlreadyRegistered cd.Code = 1000 + iota
	// MetricNotFound indicates the requested metric was not found
	MetricNotFound
	// InvalidMetricType indicates an invalid metric type was specified
	InvalidMetricType
	// InvalidMetricValue indicates an invalid metric value was provided
	InvalidMetricValue
	// CollectorNotInitialized indicates the collector is not properly initialized
	CollectorNotInitialized
	// RegistryNotInitialized indicates the registry is not properly initialized
	RegistryNotInitialized
	// ProviderAlreadyRegistered indicates a provider with the same name is already registered
	ProviderAlreadyRegistered
	// ProviderNotFound indicates the requested provider was not found
	ProviderNotFound
	// InvalidConfiguration indicates invalid monitoring configuration
	InvalidConfiguration
	// ExportFailed indicates metric export failed
	ExportFailed
	// SamplingDisabled indicates sampling is disabled for this metric
	SamplingDisabled
	// BufferFull indicates the metric buffer is full
	BufferFull
	// OperationTimeout indicates the monitoring operation timed out
	OperationTimeout
	// ResourceExhausted indicates monitoring resources are exhausted
	ResourceExhausted
)

// NewError creates a new monitoring error
func NewError(code cd.Code, message string) *cd.Error {
	return cd.NewError(code, message)
}

// NewMetricAlreadyRegisteredError creates an error for duplicate metric registration
func NewMetricAlreadyRegisteredError(metricName string) *cd.Error {
	return NewError(MetricAlreadyRegistered, "metric '"+metricName+"' is already registered")
}

// NewMetricNotFoundError creates an error for missing metric
func NewMetricNotFoundError(metricName string) *cd.Error {
	return NewError(MetricNotFound, "metric '"+metricName+"' not found")
}

// NewInvalidMetricTypeError creates an error for invalid metric type
func NewInvalidMetricTypeError(metricType string) *cd.Error {
	return NewError(InvalidMetricType, "invalid metric type: "+metricType)
}

// NewInvalidMetricValueError creates an error for invalid metric value
func NewInvalidMetricValueError(metricName string, value float64) *cd.Error {
	return NewError(InvalidMetricValue, "invalid value "+stringifyFloat(value)+" for metric '"+metricName+"'")
}

// NewCollectorNotInitializedError creates an error for uninitialized collector
func NewCollectorNotInitializedError() *cd.Error {
	return NewError(CollectorNotInitialized, "collector is not initialized")
}

// NewRegistryNotInitializedError creates an error for uninitialized registry
func NewRegistryNotInitializedError() *cd.Error {
	return NewError(RegistryNotInitialized, "registry is not initialized")
}

// NewProviderAlreadyRegisteredError creates an error for duplicate provider registration
func NewProviderAlreadyRegisteredError(providerName string) *cd.Error {
	return NewError(ProviderAlreadyRegistered, "provider '"+providerName+"' is already registered")
}

// NewProviderNotFoundError creates an error for missing provider
func NewProviderNotFoundError(providerName string) *cd.Error {
	return NewError(ProviderNotFound, "provider '"+providerName+"' not found")
}

// NewInvalidConfigurationError creates an error for invalid configuration
func NewInvalidConfigurationError(field string, value interface{}, message string) *cd.Error {
	return NewError(InvalidConfiguration, "invalid configuration: "+field+"="+stringify(value)+": "+message)
}

// NewExportFailedError creates an error for failed metric export
func NewExportFailedError(reason string) *cd.Error {
	return NewError(ExportFailed, "metric export failed: "+reason)
}

// NewSamplingDisabledError creates an error for disabled sampling
func NewSamplingDisabledError() *cd.Error {
	return NewError(SamplingDisabled, "sampling is disabled for this metric")
}

// NewBufferFullError creates an error for full buffer
func NewBufferFullError() *cd.Error {
	return NewError(BufferFull, "metric buffer is full")
}

// NewOperationTimeoutError creates an error for operation timeout
func NewOperationTimeoutError(operation string) *cd.Error {
	return NewError(OperationTimeout, "operation '"+operation+"' timed out")
}

// NewResourceExhaustedError creates an error for exhausted resources
func NewResourceExhaustedError(resource string) *cd.Error {
	return NewError(ResourceExhausted, resource+" resources exhausted")
}

// Helper functions for string conversion
func stringify(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "number"
	case float32, float64:
		return stringifyFloat(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case time.Duration:
		return val.String()
	default:
		return "unknown"
	}
}

func stringifyFloat(v interface{}) string {
	switch val := v.(type) {
	case float32:
		return stringifyFloat64(float64(val))
	case float64:
		return stringifyFloat64(val)
	default:
		return "unknown"
	}
}

func stringifyFloat64(v float64) string {
	// Simple implementation - in production you'd use proper formatting
	if v == float64(int64(v)) {
		return stringifyInt64(int64(v))
	}
	return "float"
}

func stringifyInt64(v int64) string {
	// Simple implementation
	if v == 0 {
		return "0"
	}
	return "number"
}

// IsMonitoringError checks if an error is a monitoring-specific error
func IsMonitoringError(err *cd.Error) bool {
	if err == nil {
		return false
	}
	code := err.Code
	return code >= MetricAlreadyRegistered && code <= ResourceExhausted
}

// GetErrorCode returns the error code as a string for logging/debugging
func GetErrorCode(err *cd.Error) string {
	if err == nil {
		return "Success"
	}

	switch err.Code {
	case MetricAlreadyRegistered:
		return "MetricAlreadyRegistered"
	case MetricNotFound:
		return "MetricNotFound"
	case InvalidMetricType:
		return "InvalidMetricType"
	case InvalidMetricValue:
		return "InvalidMetricValue"
	case CollectorNotInitialized:
		return "CollectorNotInitialized"
	case RegistryNotInitialized:
		return "RegistryNotInitialized"
	case ProviderAlreadyRegistered:
		return "ProviderAlreadyRegistered"
	case ProviderNotFound:
		return "ProviderNotFound"
	case InvalidConfiguration:
		return "InvalidConfiguration"
	case ExportFailed:
		return "ExportFailed"
	case SamplingDisabled:
		return "SamplingDisabled"
	case BufferFull:
		return "BufferFull"
	case OperationTimeout:
		return "OperationTimeout"
	case ResourceExhausted:
		return "ResourceExhausted"
	default:
		return "Unknown"
	}
}
