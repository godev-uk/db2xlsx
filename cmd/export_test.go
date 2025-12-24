package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInTableIncludes(t *testing.T) {
	assert.True(t, inTableIncludes([]string{}, ""))
	assert.True(t, inTableIncludes([]string{"foo", "bar"}, "foo"))
	assert.True(t, inTableIncludes([]string{"foo", "bar"}, "bar"))
	assert.False(t, inTableIncludes([]string{"foo", "bar"}, "foot"))
}
