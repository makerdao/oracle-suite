package some

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSomeFunc(t *testing.T) {
	assert.Equal(t, "test", SomeFunc("test"))
}
