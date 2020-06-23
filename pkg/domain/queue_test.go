package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryTaskQueue_Empty(t *testing.T) {
	q := NewMemoryTaskQueue()

	select {
	case <-q.Chan():
		t.Error("nothing should be in empty queue")
	default:
		// success
	}
}

func TestMemoryTaskQueue_Close(t *testing.T) {
	q := NewMemoryTaskQueue()
	q.Close()

	select {
	case v, ok := <-q.Chan():
		assert.Nil(t, v)
		assert.False(t, ok)
	default:
		t.Error("channel should not be empty")
	}
}

func TestMemoryTaskQueue_Push(t *testing.T) {
	task := &Task{}

	q := NewMemoryTaskQueue()
	q.Push(task)

	//v, ok := <-q.Chan()
	//assert.Equal(t, task, v)
	//assert.True(t, ok)

	select {
	case v, ok := <-q.Chan():
		assert.Equal(t, task, v)
		assert.True(t, ok)
	default:
		t.Error("channel should not be empty")
	}
}
