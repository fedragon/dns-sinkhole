package dns

import (
	"bufio"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fedragon/sinkhole/internal/blacklist"
)

func TestSinkhole(t *testing.T) {
	s := NewSinkhole(slog.Default())
	file, err := os.Open("./test-hosts")
	assert.NoError(t, err)
	defer file.Close()

	for domain := range blacklist.Parse(bufio.NewScanner(file)) {
		assert.NoError(t, s.Register(domain))
	}

	for domain := range blacklist.Parse(bufio.NewScanner(file)) {
		assert.True(t, s.Contains(domain))
	}

	assert.False(t, s.Contains("federico.is"))
	assert.False(t, s.Contains("github.com"))
}
