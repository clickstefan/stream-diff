package patterndetection

import (
	"data-comparator/internal/pkg/config"
	"fmt"
	"regexp"
	"strings"
)

// OfflineDetector uses built-in pattern recognition for detecting regex patterns.
type OfflineDetector struct {
	config *config.OfflineModelConfig
}

// NewOfflineDetector creates a new offline pattern detector.
func NewOfflineDetector(cfg *config.OfflineModelConfig) (*OfflineDetector, error) {
	return &OfflineDetector{config: cfg}, nil
}

// DetectPatterns analyzes field values and generates appropriate regex patterns.
func (d *OfflineDetector) DetectPatterns(fieldName string, fieldType string, values []interface{}) ([]Matcher, error) {
	if len(values) == 0 {
		return []Matcher{}, nil
	}

	var matchers []Matcher
	
	// Convert values to strings for pattern analysis
	stringValues := make([]string, 0, len(values))
	for _, val := range values {
		if val != nil {
			stringValues = append(stringValues, fmt.Sprintf("%v", val))
		}
	}

	if len(stringValues) == 0 {
		return []Matcher{}, nil
	}

	// Apply built-in pattern detection logic
	if pattern := d.detectEmailPattern(stringValues); pattern != "" {
		matchers = append(matchers, Matcher{"regex": pattern})
	} else if pattern := d.detectPhonePattern(stringValues); pattern != "" {
		matchers = append(matchers, Matcher{"regex": pattern})
	} else if pattern := d.detectURLPattern(stringValues); pattern != "" {
		matchers = append(matchers, Matcher{"regex": pattern})
	} else if pattern := d.detectIPPattern(stringValues); pattern != "" {
		matchers = append(matchers, Matcher{"regex": pattern})
	} else if pattern := d.detectUUIDPattern(stringValues); pattern != "" {
		matchers = append(matchers, Matcher{"regex": pattern})
	} else if fieldType == "numeric" {
		matchers = append(matchers, Matcher{"isNumeric": true})
	} else if fieldType == "datetime" {
		matchers = append(matchers, Matcher{"isDateTime": true})
	}

	return matchers, nil
}

// detectEmailPattern checks if values match email patterns.
func (d *OfflineDetector) detectEmailPattern(values []string) string {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	matchCount := 0
	
	for _, val := range values {
		if emailRegex.MatchString(val) {
			matchCount++
		}
	}

	// If more than 80% of values match email pattern, consider it an email field
	if float64(matchCount)/float64(len(values)) > 0.8 {
		return `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	}
	return ""
}

// detectPhonePattern checks if values match phone number patterns.
func (d *OfflineDetector) detectPhonePattern(values []string) string {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$|^\(\d{3}\)\s\d{3}-\d{4}$|^\d{3}-\d{3}-\d{4}$|^\d{10,15}$`)
	matchCount := 0
	
	// Phone numbers should be at least 7 digits and contain some specific formatting patterns
	for _, val := range values {
		// Skip short numeric values that are likely not phone numbers
		if len(val) < 7 {
			continue
		}
		
		// Check for phone-like patterns but exclude simple numbers like ages
		if phoneRegex.MatchString(val) {
			// Additional check: if all values are short numbers (like 2 digits), probably not phone numbers
			if len(val) <= 3 {
				continue
			}
			matchCount++
		}
	}

	// Require higher threshold and longer values for phone detection
	if float64(matchCount)/float64(len(values)) > 0.8 && len(values) > 0 {
		// Double-check that most values look like phone numbers (longer than typical ages/IDs)
		longValueCount := 0
		for _, val := range values {
			if len(val) >= 7 {
				longValueCount++
			}
		}
		if float64(longValueCount)/float64(len(values)) > 0.5 {
			return `^\+?[1-9]\d{1,14}$|^\(\d{3}\)\s\d{3}-\d{4}$|^\d{3}-\d{3}-\d{4}$`
		}
	}
	return ""
}

// detectURLPattern checks if values match URL patterns.
func (d *OfflineDetector) detectURLPattern(values []string) string {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	matchCount := 0
	
	for _, val := range values {
		if urlRegex.MatchString(val) {
			matchCount++
		}
	}

	if float64(matchCount)/float64(len(values)) > 0.8 {
		return `^https?://[^\s/$.?#].[^\s]*$`
	}
	return ""
}

// detectIPPattern checks if values match IP address patterns.
func (d *OfflineDetector) detectIPPattern(values []string) string {
	ipRegex := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	matchCount := 0
	
	for _, val := range values {
		if ipRegex.MatchString(val) {
			matchCount++
		}
	}

	if float64(matchCount)/float64(len(values)) > 0.8 {
		return `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`
	}
	return ""
}

// detectUUIDPattern checks if values match UUID patterns.
func (d *OfflineDetector) detectUUIDPattern(values []string) string {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	matchCount := 0
	
	for _, val := range values {
		if uuidRegex.MatchString(strings.ToLower(val)) {
			matchCount++
		}
	}

	if float64(matchCount)/float64(len(values)) > 0.8 {
		return `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	}
	return ""
}