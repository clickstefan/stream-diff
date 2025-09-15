package patterndetection

import (
	"bytes"
	"data-comparator/internal/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// OnlineDetector uses external AI APIs (Claude/Anthropic) for pattern detection.
type OnlineDetector struct {
	config     *config.OnlineAPIConfig
	httpClient *http.Client
}

// NewOnlineDetector creates a new online pattern detector.
func NewOnlineDetector(cfg *config.OnlineAPIConfig) (*OnlineDetector, error) {
	if cfg == nil {
		return nil, fmt.Errorf("online API configuration is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for online mode")
	}
	if cfg.Provider == "" {
		cfg.Provider = "claude" // Default to Claude
	}

	return &OnlineDetector{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// DetectPatterns uses AI API to analyze field values and generate regex patterns.
func (d *OnlineDetector) DetectPatterns(fieldName string, fieldType string, values []interface{}) ([]Matcher, error) {
	if len(values) == 0 {
		return []Matcher{}, nil
	}

	// Sample values for AI analysis (limit to avoid huge API calls)
	sampleValues := d.sampleValues(values, 10)
	if len(sampleValues) == 0 {
		return []Matcher{}, nil
	}

	prompt := d.buildPrompt(fieldName, fieldType, sampleValues)
	
	var response string
	var err error
	
	switch d.config.Provider {
	case "claude", "anthropic":
		response, err = d.callClaudeAPI(prompt)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", d.config.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to call AI API: %w", err)
	}

	return d.parseAIResponse(response, fieldType)
}

// sampleValues extracts a representative sample of values for AI analysis.
func (d *OnlineDetector) sampleValues(values []interface{}, maxSamples int) []string {
	stringValues := make([]string, 0, len(values))
	seen := make(map[string]bool)
	
	for _, val := range values {
		if val != nil {
			str := fmt.Sprintf("%v", val)
			if !seen[str] && len(stringValues) < maxSamples {
				stringValues = append(stringValues, str)
				seen[str] = true
			}
		}
	}
	
	return stringValues
}

// buildPrompt creates a prompt for the AI to analyze field patterns.
func (d *OnlineDetector) buildPrompt(fieldName string, fieldType string, sampleValues []string) string {
	return fmt.Sprintf(`Analyze the following data field and generate appropriate regex patterns if applicable.

Field Name: %s
Field Type: %s
Sample Values:
%s

Please analyze these values and determine if they follow a specific pattern that can be captured with a regex. 
If a clear pattern exists (like email addresses, phone numbers, URLs, UUIDs, etc.), provide ONLY the regex pattern.
If no clear pattern exists, respond with "NO_PATTERN".

Rules:
1. Only return a single regex pattern or "NO_PATTERN"
2. The pattern should match at least 80%% of the provided samples
3. Focus on common data patterns: emails, phones, URLs, IDs, codes, etc.
4. Do not include explanations, just the regex or "NO_PATTERN"

Response:`, fieldName, fieldType, strings.Join(sampleValues, "\n"))
}

// callClaudeAPI makes a request to Claude/Anthropic API.
func (d *OnlineDetector) callClaudeAPI(prompt string) (string, error) {
	endpoint := d.config.Endpoint
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1/messages"
	}

	model := d.config.Model
	if model == "" {
		model = "claude-3-haiku-20240307" // Use fastest, cheapest model for pattern detection
	}

	requestBody := map[string]interface{}{
		"model":      model,
		"max_tokens": 100, // We only need a short response
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", d.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return strings.TrimSpace(response.Content[0].Text), nil
}

// parseAIResponse parses the AI response and creates appropriate matchers.
func (d *OnlineDetector) parseAIResponse(response, fieldType string) ([]Matcher, error) {
	response = strings.TrimSpace(response)
	
	if response == "NO_PATTERN" || response == "" {
		// Fall back to basic type-based matchers
		var matchers []Matcher
		if fieldType == "numeric" {
			matchers = append(matchers, Matcher{"isNumeric": true})
		} else if fieldType == "datetime" {
			matchers = append(matchers, Matcher{"isDateTime": true})
		}
		return matchers, nil
	}

	// Validate that the response is a valid regex pattern
	_, err := regexp.Compile(response)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern from AI: %s, error: %w", response, err)
	}

	return []Matcher{{"regex": response}}, nil
}