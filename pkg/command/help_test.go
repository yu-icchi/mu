package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		command  string
		expected *Help
	}{
		{
			command:  "mu help",
			expected: &Help{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			t.Parallel()
			cmd, err := Parse(tt.command)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, cmd)
		})
	}
}
