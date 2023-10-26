package dns

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSinkhole_Register(t *testing.T) {
	s := NewSinkhole(slog.Default())

	assert.NoError(t, s.Register("www.federico.is"))
	assert.NoError(t, s.Register("www.github.is"))

	assert.True(t, s.Contains("www.federico.is"))
	assert.True(t, s.Contains("www.github.is"))
}
