package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomain(t *testing.T) {
	d, err := NewDomain("www.federico.is")

	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.True(t, d.Contains("www.federico.is"))
	assert.False(t, d.Contains("test.federico.is"))
	assert.False(t, d.Contains("federico.is"))
	assert.False(t, d.Contains("is"))

	assert.Error(t, d.Register("github.com"))

	assert.NoError(t, d.Register("federico.is"))

	assert.True(t, d.Contains("www.federico.is"))
	assert.True(t, d.Contains("federico.is"))
	assert.False(t, d.Contains("is"))

	assert.NoError(t, d.Register("test.federico.is"))

	assert.True(t, d.Contains("www.federico.is"))
	assert.True(t, d.Contains("test.federico.is"))
	assert.True(t, d.Contains("federico.is"))
	assert.False(t, d.Contains("is"))
}
