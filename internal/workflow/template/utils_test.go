package template

import (
	"testing"
)

func TestParseJSONSnippet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Markdown JSON block",
			input:    "Here is the result:\n```json\n{\"elements\": []}\n```\nHope it helps.",
			expected: "{\"elements\": []}",
		},
		{
			name:     "Generic code block",
			input:    "Result:\n```\n{\"elements\": [{\"id\": 1}]}\n```",
			expected: "{\"elements\": [{\"id\": 1}]}",
		},
		{
			name:     "No block",
			input:    "{\"elements\": []}",
			expected: "{\"elements\": []}",
		},
		{
			name:     "Incomplete block",
			input:    "```json\n{\"elements\": []}",
			expected: "{\"elements\": []}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseJSONSnippet(tt.input)
			if got != tt.expected {
				t.Errorf("parseJSONSnippet() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetShapeCSS(t *testing.T) {
	// Since getShapeCSS uses internal unidoc/unioffice types, we might need minimalist mocks or rely on exported logic.
	// But getShapeCSS is small, we can test it if we can construct the structs.
	// For simplicity, we can test with nil or empty structs first.
	got := getShapeCSS(nil)
	if got != "" {
		t.Errorf("getShapeCSS(nil) should be empty string, got %s", got)
	}
}

func TestParseClusterJSONSnippet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Markdown JSON block",
			input:    "Here is the cluster result:\n```json\n{\"layouts\": []}\n```",
			expected: "{\"layouts\": []}",
		},
		{
			name:     "No block",
			input:    "{\"layouts\": []}",
			expected: "{\"layouts\": []}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseClusterJSONSnippet(tt.input)
			if got != tt.expected {
				t.Errorf("parseClusterJSONSnippet() = %v, want %v", got, tt.expected)
			}
		})
	}
}
