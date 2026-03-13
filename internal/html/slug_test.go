package html

import "testing"

func TestSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"New York City Marathon", "new-york-city-marathon"},
		{"Brooklyn Half-Marathon", "brooklyn-half-marathon"},
		{"St. George Marathon", "st-george-marathon"},
		{"New York 10 & 5 Mile Race", "new-york-10-5-mile-race"},
		{"Some Race (2024)", "some-race-2024"},
		{"  Weird  Spacing  ", "weird-spacing"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Slug(tt.input)
			if result != tt.expected {
				t.Errorf("Slug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
