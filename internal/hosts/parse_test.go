package hosts

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_ReturnsEmptyChannel_OnEmptyInput(t *testing.T) {
	ch := Parse(bufio.NewScanner(strings.NewReader("")))
	_, ok := <-ch
	assert.False(t, ok)
}

func TestParse_OnlyReturnsHosts_AppearingAfterMarker(t *testing.T) {
	input := fmt.Sprintf(`
# start stevenblack
1.2.3.4 www.federico.is
`)
	ch := Parse(bufio.NewScanner(strings.NewReader(input)))
	domain, ok := <-ch
	assert.True(t, ok)
	assert.Equal(t, "www.federico.is", domain.Domain)
}

func TestParse_IgnoresCommentedLines(t *testing.T) {
	input := fmt.Sprintf(`
# start stevenblack
# 1.2.3.4 www.federico.is
`)
	ch := Parse(bufio.NewScanner(strings.NewReader(input)))
	_, ok := <-ch
	assert.False(t, ok)
}

func TestParse_IgnoresMalformedLines(t *testing.T) {
	input := fmt.Sprintf(`
# start stevenblack
1.2.3.4
`)
	ch := Parse(bufio.NewScanner(strings.NewReader(input)))
	_, ok := <-ch
	assert.False(t, ok)
}

func TestParse_IsUnaffectedByTrailingComments(t *testing.T) {
	input := fmt.Sprintf(`
# start stevenblack
1.2.3.4 www.federico.is # trailing comment
`)
	ch := Parse(bufio.NewScanner(strings.NewReader(input)))
	domain, ok := <-ch
	assert.True(t, ok)
	assert.Equal(t, "www.federico.is", domain.Domain)
}
