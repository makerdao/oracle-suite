package some

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSomeFunc(t *testing.T) {
	assert.Equal(t, "test", SomeFunc("test"))
}
