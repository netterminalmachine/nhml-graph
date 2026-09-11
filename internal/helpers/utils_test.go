package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SanitizeMigrationName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string with spaces",
			input:    "add foo table",
			expected: "add-foo-table",
		},
		{
			name:     "no spaces",
			input:    "AddFooTable",
			expected: "AddFooTable",
		},
		{
			name:     "with special characters",
			input:    "ajouter #$table spéciale",
			expected: "ajouter-table-speciale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SanitizeMigrationName(tt.input))
		})
	}
}
