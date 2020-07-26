package vector

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendAndGet(t *testing.T) {
	v := New()
	for i := 0; i < 2000; i++ {
		v = v.Append(i + 1)
	}

	for i := 0; i < 2000; i++ {
		require.Equal(t, v.Get(i), i+1)
	}
}

func TestGet(t *testing.T) {
	require := require.New(t)

	v := New(1, 2, 3, 4, 5)
	require.Equal(1, v.Get(0))
	require.Equal(5, v.Get(-1))
	require.Equal(3, v.Get(-3))
	require.Nil(v.Get(55))
}

func TestTake(t *testing.T) {
	require.True(t, Equal(New(1, 2, 3).Take(2), New(1, 2)))
	require.True(t, Equal(New(1, 2, 3).Take(50), New(1, 2, 3)))
}

func TestDrop(t *testing.T) {
	require.True(t, Equal(New(1, 2, 3, 4).Drop(2), New(3, 4)))
	require.Equal(t, 2, len(New(1, 2, 3, 4).Drop(2).Slice()))
}

func TestSlice(t *testing.T) {
	require.Equal(t, []interface{}{1, 2, 3}, New(1, 2, 3).Slice())
}

func TestSet(t *testing.T) {
	require := require.New(t)

	v := New(1, 2, 3, 4, 5).
		Set(0, -1).
		Set(1, -2).
		Set(2, -3).
		Set(-1, -5)

	require.True(Equal(New(-1, -2, -3, 4, -5), v))

	require.Panics(func() {
		New().Set(0, 1)
	})

	require.Equal(-1, makeVector(10000).Set(0, -1).First())
}
func TestTail(t *testing.T) {
	v := New(1, 2, 3)
	require.True(t, Equal(New(2, 3), v.Tail()))
	require.Equal(t, New(), New(1).Tail())
}

func TestRange(t *testing.T) {
	require := require.New(t)

	v := New(1, 2, 3, 4, 5, 6)
	var result []interface{}
	err := v.Range(func(elem interface{}) error {
		result = append(result, elem)
		return nil
	})
	require.NoError(err)
	expected := []interface{}{1, 2, 3, 4, 5, 6}
	require.Equal(expected, result)

	result = nil
	var i int
	err = v.Range(func(elem interface{}) error {
		result = append(result, elem)
		if i == 3 {
			return ErrStop
		}
		i++
		return nil
	})
	require.NoError(err)
	expected = []interface{}{1, 2, 3, 4}
	require.Equal(expected, result)

	var someErr = fmt.Errorf("foo")
	err = v.Range(func(elem interface{}) error {
		return someErr
	})
	require.Equal(someErr, err)
}

func TestEqual(t *testing.T) {
	require.True(t, Equal(New(1, 2, 3), New(1, 2, 3)))
	require.False(t, Equal(New(1, 2), New(1, 2, 3)))
	require.False(t, Equal(New(1, 2, 4), New(1, 2, 3)))
}

func TestVectorString(t *testing.T) {
	v := New(1, 2, 3, 4, 5)
	require.Equal(t, v.String(), "[1, 2, 3, 4, 5]")
}

func TestVectorFirst(t *testing.T) {
	v := New(1, 2, 3, 4, 5)
	el := v.First()
	require.Equal(t, el, 1)
}

func TestVectorLast(t *testing.T) {
	v := New(1, 2, 3, 4, 5)
	el := v.Last()
	require.Equal(t, el, 5)
}

func TestMap(t *testing.T) {
	v := New(1, 2, 3).Map(func(x interface{}) interface{} {
		return x.(int) * x.(int)
	})

	require.True(t, Equal(v, New(1, 4, 9)))
}

func TestFilter(t *testing.T) {
	v := New(1, 2, 3).Filter(func(x interface{}) bool {
		return x.(int)%2 == 1
	})

	require.True(t, Equal(v, New(1, 3)))
}

func BenchmarkAppend(b *testing.B) {
	v10 := makeVector(10)
	v100 := makeVector(100)
	v1000 := makeVector(1000)

	fn := func(v *Vector) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v = v.Append(i)
			}
		}
	}

	b.Run("10", fn(v10))
	b.Run("100", fn(v100))
	b.Run("1000", fn(v1000))
}

func BenchmarkGet(b *testing.B) {
	v10 := makeVector(10)
	v100 := makeVector(100)
	v1000 := makeVector(1000)

	fn := func(n int, v *Vector) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x := v.Get(i)
				if i >= n {
					require.Nil(b, x)
				} else {
					require.Equal(b, x, i)
				}
			}
		}
	}

	b.Run("10", fn(10, v10))
	b.Run("100", fn(100, v100))
	b.Run("1000", fn(1000, v1000))
}

func BenchmarkSet(b *testing.B) {
	v10 := makeVector(10)
	v100 := makeVector(100)
	v1000 := makeVector(1000)

	fn := func(n int, v *Vector) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v = v.Set(i%n, i)
			}
		}
	}

	b.Run("10", fn(10, v10))
	b.Run("100", fn(100, v100))
	b.Run("1000", fn(1000, v1000))
}

func makeVector(len int) *Vector {
	v := New()
	for i := 0; i < len; i++ {
		v = v.Append(i)
	}
	return v
}
