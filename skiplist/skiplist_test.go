package skiplist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestPutGet(t *testing.T) {
	const N = 28
	l := New[int, float64](cmpInt)
	for i := 0; i < N; i += 2 {
		l.Put(i, float64(i))
	}
	for i := N - N%2 - 1; i >= 0; i -= 2 {
		l.Put(i, float64(i))
	}
	Dump(&l)
	v, exists := l.Get(N / 2)
	require.True(t, exists, "%d should exist", N/2)
	require.Equal(t, float64(int(N/2)), v)
}

func TestCC(t *testing.T) {
	l := New[int, int](cmpInt)
	eg := errgroup.Group{}
	const N = 100
	const W = 100
	for i := 0; i < W; i++ {
		i := i
		eg.Go(func() error {
			for j := 0; j < N; j++ {
				l.Put(j*i, j)
			}
			return nil
		})
	}
	eg.Wait()
	// Dump(&l)
}

func BenchmarkPut(b *testing.B) {
	l := New[int, int](cmpInt)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Put(i, i)
	}
}

func BenchmarkGet(b *testing.B) {
	const numKeys = 1e5
	l := New[int, int](cmpInt)
	for i := 0; i < numKeys; i++ {
		l.Put(i, i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		k := (i * 100) % numKeys
		_, exists := l.Get(k)
		if !exists {
			b.Fatalf("missing %v", k)
		}
	}
}

func cmpInt(a, b int) int {
	return a - b
}

// Dump dumps a list
func Dump[K, V any](l *List[K, V]) {
	for i := 0; ; i++ {
		n := l.getHead(i)
		if n == nil {
			break
		}
		fmt.Printf("level %d\n", i)
		fmt.Print("HEAD -> ")
		for n != nil {
			fmt.Printf("{%v: %v}-> ", n.key, n.value)
			n = n.getSuc(i)
		}
		fmt.Println("nil")
	}
}
