package immutable

import (
	"fmt"
	"strings"
)

// Set stores Value objects in a HAMT structure
type Set struct {
	l int   // number of entries
	m *HAMT // trie root
}

// The empty Set
var EmptySet = &Set{0, EmptyHAMT}

// Len returns the number of entries
func (s *Set) Len() int { return s.l }

// Get finds value for v. Returns nil if not found.
func (s *Set) Get(v Value) Value { return s.m.Lookup(v.Hash(), v) }

// Has returns true if v is in the set.
func (s *Set) Has(v Value) bool { return s.Get(v) != nil }

// Add returns a Set which contains v
func (s *Set) Add(v Value) *Set {
	len2 := s.l + 1
	m2 := s.m.Insert(0, v.Hash(), v, &len2)
	return &Set{len2, m2}
}

// Del returns a Set without v. If v is not found, returns the receiver.
func (s *Set) Del(v Value) *Set {
	m2 := s.m.Remove(v.Hash(), v)
	if m2 == s.m {
		return s // not found; no change
	}
	return &Set{s.l - 1, m2}
}

// Range iterates over all values by calling f(v). If f returns false, iteration stops.
// Order is by key path.
func (s *Set) Range(f func(Value) bool) {
	s.m.Range(f)
}

// String returns human-readable text in the format "{Value, Value, Value}"
func (s *Set) String() string { return stringSet(s.m) }

func stringSet(m *HAMT) string {
	var sb strings.Builder
	sb.WriteByte('{')
	first := true
	m.Range(func(v Value) bool {
		if first {
			first = false
		} else {
			sb.WriteString(", ")
		}
		fmt.Fprint(&sb, v)
		return true
	})
	sb.WriteByte('}')
	return sb.String()
}

// —————————————————————————————————————————————

// StrSet stores strings in a HAMT structure
type StrSet struct {
	l int   // number of entries
	m *HAMT // trie root
}

// The empty StrSet
var EmptyStrSet = &StrSet{0, EmptyHAMT}

// Len returns the number of entries
func (s *StrSet) Len() int { return s.l }

// Has returns true if v is in the set.
func (s *StrSet) Has(v string) bool {
	val := StrValue{StrHash(v), v}
	return s.m.Lookup(val.H, &val) != nil
}

// Add returns a StrSet which contains v
func (s *StrSet) Add(v string) *StrSet {
	val := &StrValue{StrHash(v), v}
	len2 := s.l + 1
	m2 := s.m.Insert(0, val.H, val, &len2)
	return &StrSet{len2, m2}
}

// Del returns a StrSet without s. If s is not found, returns the receiver.
func (s *StrSet) Del(v string) *StrSet {
	val := StrValue{StrHash(v), v}
	m2 := s.m.Remove(val.H, &val)
	if m2 == s.m {
		return s // not found; no change
	}
	return &StrSet{s.l - 1, m2}
}

// Range iterates over all values by calling f(v). If f returns false, iteration stops.
// Order is by key path.
func (s *StrSet) Range(f func(string) bool) {
	s.m.Range(func(v Value) bool { return f(v.(*StrValue).K) })
}

// String returns human-readable text in the format "{Value, Value, Value}"
func (s *StrSet) String() string { return stringSet(s.m) }
