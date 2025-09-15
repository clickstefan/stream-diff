package patterndetection

import (
	"data-comparator/internal/pkg/config"
	"fmt"
)

// Matcher is a flexible map to represent matcher configurations,
// e.g., {"isNumeric": true} or {"regex": "pattern"}.
type Matcher map[string]interface{}

// PatternDetector interface defines methods for detecting regex patterns in field values.
type PatternDetector interface {
	DetectPatterns(fieldName string, fieldType string, values []interface{}) ([]Matcher, error)
}

// DetectorFactory creates pattern detectors based on configuration.
type DetectorFactory struct {
	config *config.PatternDetection
}

// NewDetectorFactory creates a new detector factory with the given configuration.
func NewDetectorFactory(cfg *config.PatternDetection) *DetectorFactory {
	return &DetectorFactory{config: cfg}
}

// CreateDetector creates a pattern detector based on the configuration.
func (f *DetectorFactory) CreateDetector() (PatternDetector, error) {
	if f.config == nil || !f.config.Enabled {
		return &NoOpDetector{}, nil
	}

	switch f.config.Mode {
	case "offline":
		return NewOfflineDetector(f.config.OfflineModel)
	case "online":
		return NewOnlineDetector(f.config.OnlineAPI)
	default:
		return nil, fmt.Errorf("unsupported pattern detection mode: %s", f.config.Mode)
	}
}

// NoOpDetector is a no-operation detector that returns empty matchers.
type NoOpDetector struct{}

// DetectPatterns implements PatternDetector interface with no-op behavior.
func (d *NoOpDetector) DetectPatterns(fieldName string, fieldType string, values []interface{}) ([]Matcher, error) {
	return []Matcher{}, nil
}