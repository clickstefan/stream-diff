package patterndetection

import (
	"data-comparator/internal/pkg/config"
	"testing"
)

func TestOfflineDetector_DetectEmailPattern(t *testing.T) {
	detector := &OfflineDetector{}
	
	testCases := []struct {
		name     string
		values   []interface{}
		expected bool
	}{
		{
			name:     "valid emails",
			values:   []interface{}{"alice@example.com", "bob@test.org", "charlie@domain.net"},
			expected: true,
		},
		{
			name:     "mixed valid and invalid",
			values:   []interface{}{"alice@example.com", "not-an-email", "bob@test.org"},
			expected: false,
		},
		{
			name:     "no emails",
			values:   []interface{}{"john doe", "123456", "not an email"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchers, err := detector.DetectPatterns("email", "string", tc.values)
			if err != nil {
				t.Fatalf("DetectPatterns failed: %v", err)
			}

			hasEmailRegex := false
			for _, matcher := range matchers {
				if regex, ok := matcher["regex"]; ok {
					if regex == `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` {
						hasEmailRegex = true
						break
					}
				}
			}

			if hasEmailRegex != tc.expected {
				t.Errorf("Expected email regex detection: %v, got: %v", tc.expected, hasEmailRegex)
			}
		})
	}
}

func TestDetectorFactory_CreateDetector(t *testing.T) {
	testCases := []struct {
		name     string
		config   *config.PatternDetection
		wantType string
		wantErr  bool
	}{
		{
			name:     "disabled",
			config:   &config.PatternDetection{Enabled: false},
			wantType: "*patterndetection.NoOpDetector",
			wantErr:  false,
		},
		{
			name:     "offline mode",
			config:   &config.PatternDetection{Enabled: true, Mode: "offline"},
			wantType: "*patterndetection.OfflineDetector",
			wantErr:  false,
		},
		{
			name:     "online mode with config",
			config:   &config.PatternDetection{
				Enabled: true, 
				Mode: "online",
				OnlineAPI: &config.OnlineAPIConfig{APIKey: "test-key"},
			},
			wantType: "*patterndetection.OnlineDetector",
			wantErr:  false,
		},
		{
			name:     "online mode without API key",
			config:   &config.PatternDetection{Enabled: true, Mode: "online"},
			wantType: "",
			wantErr:  true,
		},
		{
			name:     "invalid mode",
			config:   &config.PatternDetection{Enabled: true, Mode: "invalid"},
			wantType: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factory := NewDetectorFactory(tc.config)
			detector, err := factory.CreateDetector()

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if detector == nil {
				t.Fatal("Detector is nil")
			}

			// Check type (basic type checking)
			detectorType := ""
			switch detector.(type) {
			case *NoOpDetector:
				detectorType = "*patterndetection.NoOpDetector"
			case *OfflineDetector:
				detectorType = "*patterndetection.OfflineDetector"
			case *OnlineDetector:
				detectorType = "*patterndetection.OnlineDetector"
			}

			if detectorType != tc.wantType {
				t.Errorf("Expected detector type %s, got %s", tc.wantType, detectorType)
			}
		})
	}
}

func TestNoOpDetector_DetectPatterns(t *testing.T) {
	detector := &NoOpDetector{}
	
	matchers, err := detector.DetectPatterns("test_field", "string", []interface{}{"value1", "value2"})
	if err != nil {
		t.Fatalf("DetectPatterns failed: %v", err)
	}

	if len(matchers) != 0 {
		t.Errorf("Expected empty matchers, got %d", len(matchers))
	}
}