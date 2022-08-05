package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPipeli(t *testing.T) {
	p := NewPipeline()

	assert.NotNil(t, p)
}

func TestAddSegment(t *testing.T) {
	p := NewPipeline()

	fn := func(input interface{}) interface{} {
		return input
	}

	p.AddSegment(fn)

	assert.Equal(t, 1, len(p))
}

func TestAddSegmentWithCapacity(t *testing.T) {
	p := NewPipeline()

	fn := func(input interface{}) interface{} {
		return input
	}

	p.AddSegmentWithCapacity(fn, 10)

	assert.Equal(t, 1, len(p))
}

func TestRun(t *testing.T) {
	p := NewPipeline()

	input := make(chan interface{}, 100)
	for i := 0; i < 10; i++ {
		input <- i
	}
	close(input)

	fn := func(input interface{}) interface{} {
		return input.(int) + 1
	}

	result := make([]interface{}, 0)
	end := func(input interface{}) interface{} {
		result = append(result, input)

		return nil
	}

	p.AddSegment(fn)
	p.AddSegment(fn)
	p.AddSegment(end)

	done := p.Run(input)
	_, ok := <-done
	assert.False(t, ok)

	assert.Equal(t, 10, len(result))
	for i := 0; i < 10; i++ {
		assert.Equal(t, i+2, result[i])
	}
}
