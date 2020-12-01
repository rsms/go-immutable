package immutable

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Example_strMap() {
	m := EmptyStrMap
	m1 := m.Set("Hello", 123)
	m2 := m.Set("Hello", 456).Set("Sun", 9)
	m3 := m2.Del("Hello")
	fmt.Printf("m1: %s\n", m1)
	fmt.Printf("m2: %s\n", m2)
	fmt.Printf("m3: %s\n", m3)
	// Output:
	// m1: {"Hello": 123}
	// m2: {"Sun": 9, "Hello": 456}
	// m3: {"Sun": 9}
}

func TestStrMap(t *testing.T) {
	assert := assert.New(t)
	vals := testDataColorNames

	// insert and check lookup of every version of m
	m := EmptyStrMap
	for _, sample := range vals {
		m = m.Set(sample, sample)
		assert.Equal(sample, m.Get(sample))
	}
	// t.Log(m.String())

	// length should match vals (vals contains unique strings)
	assert.Equal(len(vals), m.Len())

	// inserting the same values should not grow the map
	for _, sample := range vals {
		m = m.Set(sample, sample)
	}
	assert.Equal(len(vals), m.Len())

	// Get should return the expected value
	for _, sample := range vals {
		assert.Equal(sample, m.Get(sample))
	}

	// Del should remove entries
	for _, sample := range vals {
		// entry should exist (not deleted accidentally by past Del)
		assert.True(m.Has(sample))
		m2 := m.Del(sample)
		// should return a new version of m
		assert.NotEqual(m, m2)
		m = m2
		// entry should no longer exist in m
		assert.False(m.Has(sample))
	}
	assert.Equal(0, m.Len())
}
