package state

import (
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
