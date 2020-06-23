package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackUnpack(t *testing.T) {
	src := []byte("some data string")

	packed, err := pack(src)

	assert.NotNil(t, packed)
	assert.Nil(t, err)

	dst, err := unpack(packed)

	assert.NotNil(t, dst)
	assert.Nil(t, err)

	assert.Equal(t, src, dst, "unpacked byte string should be equal to original one")
}
