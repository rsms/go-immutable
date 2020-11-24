package immutable

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Example_strSet() {
	s1 := EmptyStrSet
	s2 := s1.Add("Robin")
	s1 = s1.Add("Anne").Add("Frank")
	s3 := s1.Del("Anne")
	fmt.Printf("s1: %s\n", s1)
	fmt.Printf("s2: %s\n", s2)
	fmt.Printf("s3: %s\n", s3)
	// Output:
	// s1: {Frank, Anne}
	// s2: {Robin}
	// s3: {Frank}
}

func TestSetInsert(t *testing.T) {
	assert := assert.New(t)
	vals := []string{
		"1",
		"2",
		"31",
		"10",
		"5",
		"1/20",
		"2/20",
		"31/20",
		"10/20",
		"5/5",
		"1/5",
		"2/5",
		"31/5",
		"10/5",
		"1/2",
		"1/2/1/2/1/2",
		"1/2/1/1/1/1",
		"1/2/1/1/1/2",
		"1/2/1/1/1/3",
		"1/2/1/2/1/1",
		"1/2/1/2/1/3",
		"1/1",
	}

	s := EmptySet
	for _, vpath := range vals {
		s = s.Add(newValue(vpath))
	}
	assert.Equal(len(vals), s.Len)

	// inserting the same values should not grow the set
	for _, vpath := range vals {
		s = s.Add(newValue(vpath))
	}
	// t.Log(s.m.String())
	assert.Equal(len(vals), s.Len)

	// Get should return the expected value
	for _, vpath := range vals {
		v := s.Get(newValue(vpath))
		if !assert.NotNil(v) {
			break
		}
		assert.Equal(vpath, v.(*myValue).value)
	}
}

func TestSetCollision(t *testing.T) {
	assert := assert.New(t)
	s := EmptySet

	// set non-colliding values
	s = s.Add(newValue("1/1"))
	s = s.Add(newValue("2/1"))
	s = s.Add(newValue("3/1"))
	s = s.Add(newValue("2/2"))
	// test makeBranch() key path collision resolve
	s = s.Add(newValue("2/1/1/1/2/1/1/1/1"))
	s = s.Add(newValue("2/1/1/1/2/1/1/1/2"))

	// collision
	// note: myCollidingValue.Equal() only true for same obj
	cv1 := newCollidingValue("1", "1a")
	s = s.Add(cv1)
	s = s.Add(newCollidingValue("1", "1b"))
	s = s.Add(newCollidingValue("1", "1c"))
	s = s.Add(newCollidingValue("1", "1d"))
	// t.Log(s.m.String())
	assert.Equal(10, s.Len)
	assert.Equal(cv1, s.Get(cv1))
}

func TestSetDeleteCollision(t *testing.T) {
	assert := assert.New(t)
	cv := []*myCollidingValue{
		newCollidingValue("1", "1a"),
		newCollidingValue("1", "1b"),
		newCollidingValue("1", "1c"),
		newCollidingValue("1", "1d"),
	}
	s := EmptySet
	s = s.Add(newValue("1/1"))
	s = s.Add(newValue("2/1"))
	s = s.Add(newValue("2/2"))
	s = s.Add(newValue("3/1"))
	for _, v := range cv {
		s = s.Add(v)
		if !assert.Equal(v, s.Get(v)) {
			break
		}
	}
	assert.Equal(4+len(cv), s.Len)
	// t.Log(s.m.String())
	for _, v := range cv {
		s = s.Del(v)
		if !assert.Nil(s.Get(v), v.value) {
			break
		}
	}
	assert.Equal(4, s.Len)
}

func TestSetDeleteDeepBranches(t *testing.T) {
	assert := assert.New(t)
	s := EmptySet

	s = s.Add(newValue("1/1"))
	s = s.Add(newValue("1/2"))
	s = s.Add(newValue("1/3"))
	assert.NotNil(s.Get(newValue("1/2")))
	// t.Log(s.m.String())

	s = s.Del(newValue("1/2"))
	// t.Log(s.m.String())
	assert.Equal(s.Len, 2)
	assert.NotNil(s.Get(newValue("1/1")))
	assert.Nil(s.Get(newValue("1/2")))
	assert.NotNil(s.Get(newValue("1/3")))

	s = s.Del(newValue("1/3"))
	// t.Log(s.m.String())
	assert.Equal(s.Len, 1)
	assert.NotNil(s.Get(newValue("1/1")))
	assert.Nil(s.Get(newValue("1/2")))
	assert.Nil(s.Get(newValue("1/3")))

	s = s.Del(newValue("1/1"))
	// t.Log(s.m.String())
	assert.Equal(s.Len, 0)
	assert.Nil(s.Get(newValue("1/1")))
	assert.Nil(s.Get(newValue("1/2")))
	assert.Nil(s.Get(newValue("1/3")))

	// deeper branches
	s = s.Add(newValue("1/1"))
	s = s.Add(newValue("1/2/1/1/1/1"))
	s = s.Add(newValue("1/2/1/1/1/2"))
	s = s.Add(newValue("1/2/1/1/1/3"))
	s = s.Add(newValue("1/2/1/2/1/1"))
	s = s.Add(newValue("1/2/1/2/1/2"))
	s = s.Add(newValue("1/2/1/2/1/3"))
	assert.Equal(s.Len, 7)
	assert.NotNil(s.Get(newValue("1/1")))
	assert.NotNil(s.Get(newValue("1/2/1/1/1/1")))
	assert.NotNil(s.Get(newValue("1/2/1/1/1/2")))
	assert.NotNil(s.Get(newValue("1/2/1/1/1/3")))
	assert.NotNil(s.Get(newValue("1/2/1/2/1/1")))
	assert.NotNil(s.Get(newValue("1/2/1/2/1/2")))
	assert.NotNil(s.Get(newValue("1/2/1/2/1/3")))
	// t.Log(s.m.String())

	// Del() for a non-existing value should yield the same trie (no change)
	s2 := s.Del(newValue("9"))
	assert.Equal(s, s2)

	// test successful delete
	expectedLen := s.Len
	for _, val := range []string{
		"1/2/1/2/1/2",
		"1/2/1/1/1/1",
		"1/2/1/1/1/2",
		"1/2/1/1/1/3",
		"1/2/1/2/1/1",
		"1/2/1/2/1/3",
		"1/1",
	} {
		s = s.Del(newValue(val))
		expectedLen--
		assert.Equal(expectedLen, s.Len)
		if !assert.Nil(s.Get(newValue(val))) {
			break
		}
		// t.Log(s.m.String())
	}
}

func TestSetDeleteRootValues(t *testing.T) {
	assert := assert.New(t)
	// test deletion of values on root.
	// del first, middle and last. Then all.
	vals := []string{"1", "2", "3"}
	for _, delval := range vals {
		s := EmptySet
		for _, val := range vals {
			s = s.Add(newValue(val))
		}
		// del one
		assert.NotNil(s.Get(newValue(delval)))
		s = s.Del(newValue(delval))
		s2 := s.Del(newValue(delval + "/9")) // non-existing
		assert.Equal(s, s2)
		expectLen := len(vals) - 1
		assert.Equal(s.Len, expectLen)
		assert.Nil(s.Get(newValue(delval)))
		// del rest
		for _, delval2 := range vals {
			if delval == delval2 {
				continue
			}
			assert.NotNil(s.Get(newValue(delval2)))
			s = s.Del(newValue(delval2))
			expectLen--
			assert.Equal(s.Len, expectLen)
			assert.Nil(s.Get(newValue(delval2)))
		}
	}
}

func TestSetIteration(t *testing.T) {
	assert := assert.New(t)
	s := EmptySet

	// These test values should be ordered in this list in the same order they
	// are expected during iteration.
	// Note that for collisions, the order of insertion is reflected during iteration.
	vals := []Value{
		newCollidingValue("1", "1a"),
		newCollidingValue("1", "1b"),
		newCollidingValue("1", "1c"),
		newCollidingValue("1", "1d"),
		newValue("1/1"),
		newValue("2/1"),
		newValue("2/1/1/1/2/1/1/1/1"),
		newValue("2/1/1/1/2/1/1/1/2"),
		newValue("2/2"),
		newValue("3/1"),
		newValue("3/2"),
		newValue("3/2/4"),
		newValue("3/2/9"),
		newValue("3/3"),
	}
	for _, v := range vals {
		s = s.Add(v)
	}
	// t.Log(s.m.String())
	// t.Log(s.String())
	i := 0
	s.Range(func(v Value) bool {
		// t.Logf("#%d => %v", i, v)
		expectedVal := vals[i]
		assert.NotNil(v)
		assert.Equal(expectedVal, v)
		i++
		return true
	})
}

// ---------------------------------------------------------------------------
// Benchmarks

func makeTestData(b *testing.B, count int) []myValue {
	// use a pseudo-random number generator so that we get the same results every run.
	if b != nil {
		b.Helper()
	}
	r := rand.New(rand.NewSource(0))
	v := make([]myValue, count)
	keys := make(map[uint]bool)
	for i := 0; i < count; i++ {
		key := uint(hashFNV1aUint32(r.Uint32()))
		if _, ok := keys[key]; ok {
			// duplicate
			i--
			continue
		}
		keys[key] = true
		v[i].key = key
		v[i].value = FmtKey(key)
	}
	return v
}

// Ensure the code that is run for benchmarks is correct
func TestSetBenchmarkData(t *testing.T) {
	assert := assert.New(t)
	testData := makeTestData(nil, 2000)
	m := EmptySet
	for i := 0; i < len(testData); i++ {
		v := &testData[i]
		m = m.Add(v)
		found := m.Get(v)
		assert.Equal(v, found, "get")
		assert.Equal(i+1, m.Len, "len")
	}
	// insert a second time should not grow the set
	for i := 0; i < len(testData); i++ {
		v := &testData[i]
		m = m.Add(v)
	}
	assert.Equal(len(testData), m.Len, "final len")

	for i := 0; i < len(testData); i++ {
		v := &testData[i]
		found := m.Get(v).(*myValue)
		assert.Equal(v, found, "final get")
	}
}

func benchmarkHamtInsert(b *testing.B, count int) {
	testData := makeTestData(b, count)
	var s *Set
	for n := 0; n < b.N; n++ {
		if n%count == 0 {
			s = EmptySet
		}
		s = s.Add(&testData[n%count])
	}
}

func benchmarkGoMapInsert(b *testing.B, count int) {
	testData := makeTestData(b, count)
	var m map[string]*myValue
	for n := 0; n < b.N; n++ {
		if n%count == 0 {
			m = make(map[string]*myValue)
		}
		v := &testData[n%count]
		m[v.value] = v
	}
}

func benchmarkHamtLookup(b *testing.B, count int) {
	testData := makeTestData(b, count)
	t := EmptySet
	func() {
		b.Helper()
		for i := 0; i < count; i++ {
			t = t.Add(&testData[i])
		}
	}()
	for n := 0; n < b.N; n++ {
		_ = t.Get(&testData[n%len(testData)])
	}
}

func benchmarkGoMapLookup(b *testing.B, count int) {
	testData := makeTestData(b, count)
	m := make(map[string]*myValue, count)
	func() {
		b.Helper()
		for i := 0; i < count; i++ {
			v := &testData[i]
			m[v.value] = v
		}
	}()
	for n := 0; n < b.N; n++ {
		_ = m[testData[n%len(testData)].value]
	}
}

func BenchmarkHamtLookup_10_(b *testing.B)    { benchmarkHamtLookup(b, 10) }
func BenchmarkHamtLookup_100_(b *testing.B)   { benchmarkHamtLookup(b, 100) }
func BenchmarkHamtLookup_1000_(b *testing.B)  { benchmarkHamtLookup(b, 1000) }
func BenchmarkHamtLookup_10000_(b *testing.B) { benchmarkHamtLookup(b, 10000) }

func BenchmarkGoMapLookup_10_(b *testing.B)    { benchmarkGoMapLookup(b, 10) }
func BenchmarkGoMapLookup_100_(b *testing.B)   { benchmarkGoMapLookup(b, 100) }
func BenchmarkGoMapLookup_1000_(b *testing.B)  { benchmarkGoMapLookup(b, 1000) }
func BenchmarkGoMapLookup_10000_(b *testing.B) { benchmarkGoMapLookup(b, 10000) }

func BenchmarkHamtInsert_10_(b *testing.B)   { benchmarkHamtInsert(b, 10) }
func BenchmarkHamtInsert_25_(b *testing.B)   { benchmarkHamtInsert(b, 25) }
func BenchmarkHamtInsert_50_(b *testing.B)   { benchmarkHamtInsert(b, 50) }
func BenchmarkHamtInsert_100_(b *testing.B)  { benchmarkHamtInsert(b, 100) }
func BenchmarkHamtInsert_1000_(b *testing.B) { benchmarkHamtInsert(b, 1000) }
func BenchmarkHamtInsert_5000_(b *testing.B) { benchmarkHamtInsert(b, 5000) }

func BenchmarkGoMapInsert_10_(b *testing.B)   { benchmarkGoMapInsert(b, 10) }
func BenchmarkGoMapInsert_25_(b *testing.B)   { benchmarkGoMapInsert(b, 25) }
func BenchmarkGoMapInsert_50_(b *testing.B)   { benchmarkGoMapInsert(b, 50) }
func BenchmarkGoMapInsert_100_(b *testing.B)  { benchmarkGoMapInsert(b, 100) }
func BenchmarkGoMapInsert_1000_(b *testing.B) { benchmarkGoMapInsert(b, 1000) }
func BenchmarkGoMapInsert_5000_(b *testing.B) { benchmarkGoMapInsert(b, 5000) }

/*func debugBits() {
  logf("——————————————————————————————————————————————————————————————————————")
  logf("intSize:      %d", intSize)
  logf("hamtBranches: %d, hamtBits %d", hamtBranches, hamtBits)
  logf("hamtMask:     %s", fmtbmap(hamtMask))
  logf("offs test:    %d", (64 >> hamtBits) << hamtBits)

  // bitmap index:        20        10      3 10
  bmap := uint(0b0000000000100000000010000001011)
  // bucket index:         4         3      2 10

  isBitSet := func (bmap, index uint) bool {
    // returns true if index bit is set in bmap
    return bmap & (uint(1) << index) != 0
  }

  logf("")
  logf("bmap %s", fmtbmap(bmap))
  logf("10?  => %v", isBitSet(bmap, 10))
  logf("15?  => %v", isBitSet(bmap, 15))
  logf("20?  => %v", isBitSet(bmap, 20))
  logf("i 10 => %v", bitindex1(bmap, 10))
  logf("i 20 => %v", bitindex1(bmap, 20))

  index := uint(10)

  bitpos := uint(1) << index
  logf("bmap   %s", fmtbmap(bmap))
  logf("bitpos %s", fmtbmap(bitpos) )
  logf("       %d", bitpos )
  logf("set?   %v", bmap & bitpos != 0 )

  logf("")
  logf("A    %s", fmtbmap(bmap & (uintMax ^ (uintMax << index))) )
  logf("B    %s", fmtbmap(bmap & (uintMax >> (intSize - index))) )

  // uint64_t v;       // Compute the rank (bits set) in v from the MSB to pos.
  // unsigned int pos; // Bit position to count bits upto.
  // uint64_t r;       // Resulting rank of bit at pos goes here.
  //
  // // Shift out bits after given position.
  // r = v >> (sizeof(v) * CHAR_BIT - pos);
  logf("")
  logf("a    %s", fmtbmap( bmap >> (intSize - index) ))

  logf("")
  // skm := uint(bitpos * hamtBits) - 1
  skm := uintMax >> (intSize - index)  // sub key mask
  logf("skm  %s", fmtbmap(skm) )
  logf("     %s", fmtbmap(bmap & skm) )
  logf("idx  %d", bits.OnesCount(bmap & skm) )
  logf("")
}*/
