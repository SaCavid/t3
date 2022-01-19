package merge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const size = 1000000

func TestMergesort(t *testing.T) {
	s := []int{5, 8, 9, 5, 0, 10, 1, 6}
	mergesort(s)
	assert.Equal(t, []int{0, 1, 5, 5, 6, 8, 9, 10}, s)
}

func TestParallelMergesort(t *testing.T) {
	s := []int{5, 8, 9, 5, 0, 10, 1, 6}
	ParallelMergesort(s)
	assert.Equal(t, []int{0, 1, 5, 5, 6, 8, 9, 10}, s)
}

func BenchmarkMergesort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := random(size)
		b.StartTimer()
		mergesort(s)
		b.StopTimer()
	}
}

func BenchmarkParallelMergesort2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := random(size)
		b.StartTimer()
		ParallelMergesort(s)
		b.StopTimer()
	}
}
