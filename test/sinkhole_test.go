package test

import (
	"bufio"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fedragon/sinkhole/internal/dns"
	"github.com/fedragon/sinkhole/internal/hosts"
)

func TestSinkhole(t *testing.T) {
	s := dns.NewSinkhole(slog.Default())
	file, err := os.Open("./test-hosts")
	assert.NoError(t, err)
	defer file.Close()

	for line := range hosts.Parse(bufio.NewScanner(file)) {
		assert.NoError(t, line.Err)
		s.Register(line.Domain)
	}

	for line := range hosts.Parse(bufio.NewScanner(file)) {
		assert.NoError(t, line.Err)
		assert.True(t, s.Contains(line.Domain))
	}

	assert.False(t, s.Contains("federico.is"))
	assert.False(t, s.Contains("github.com"))
}
