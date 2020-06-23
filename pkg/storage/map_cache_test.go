package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yurykabanov/scraper/pkg/domain"
)

func TestMapCache(t *testing.T) {
	cache := NewMapCache()

	result := &domain.Result{}

	v, err := cache.Get("whatever")

	assert.Nil(t, v)
	assert.Nil(t, err)

	err = cache.Put("whatever", result)

	assert.Nil(t, err)

	v, err= cache.Get("whatever")

	assert.Equal(t, result, v)
	assert.Nil(t, err)
}
