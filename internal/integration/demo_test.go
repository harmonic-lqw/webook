package integration

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDemo(t *testing.T) {
	val1 := 2
	t.Log("断点后仍能打印Log", val1)
	val2 := 3
	assert.Equal(t, 4, val2)
}
