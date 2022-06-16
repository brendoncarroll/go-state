package state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteSpan(t *testing.T) {
	t.Run("AllLt", func(t *testing.T) {
		span := ByteSpan{
			Begin: []byte("b"),
			End:   []byte("d"),
		}

		assert.True(t, span.AllLt([]byte("d")))
		assert.True(t, span.AllLt([]byte("d\x00")))
		assert.True(t, span.AllLt([]byte("z")))

		assert.False(t, span.AllLt([]byte("c")))
		assert.False(t, span.AllLt([]byte("c")))
		assert.False(t, span.AllLt(nil))
		assert.False(t, span.AllLt([]byte("b")))
	})
	t.Run("AllGt", func(t *testing.T) {
		span := ByteSpan{
			Begin: []byte("b"),
			End:   []byte("d"),
		}
		assert.True(t, span.AllGt([]byte("a")))
		assert.False(t, span.AllGt([]byte("b")))
	})
}

func ExampleSpanString() {
	fmt.Println(Span[int]{})

	a := Span[int]{}.WithLowerIncl(-5)
	fmt.Println(a)

	b := Span[int]{}.WithUpperExcl(10)
	fmt.Println(b)

	c := Span[int]{}.WithLowerIncl(2).WithUpperExcl(20)
	fmt.Println(c)

	// Output:
	// (min, max)
	// [-5, max)
	// (min, 10)
	// [2, 20)
}
