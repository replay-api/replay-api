package oracle_ocr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"navi", "navi", 0},
		{"navi", "navl", 1},     // OCR l/i confusion
		{"faze", "faz3", 1},     // OCR 3/e confusion
		{"cloud9", "cl0ud9", 1}, // OCR 0/o confusion
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			result := levenshteinDistance(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeTeamName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NAVI", "navi"},
		{"  FaZe  ", "faze"},
		{"Team Liquid", "team liquid"},
		{"G2 Esports", "g2"},
		{"Cloud9 Gaming", "cloud9"},
		{"MOUZ Team", "mouz"},
		{"  Some-Team! ", "someteam"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeTeamName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeRegex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"navi", "navi"},
		{"g2.", `g2\.`},
		{"team*", `team\*`},
		{"a[b]c", `a\[b\]c`},
		{"(test)", `\(test\)`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeRegex(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
