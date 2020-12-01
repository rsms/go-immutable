package immutable

import (
	"fmt"
	"strings"
)

// StrMap stores string keys associated with any value in a HAMT structure
type StrMap struct {
	l int   // number of entries
	m *HAMT // trie root
}

// The empty StrMap
var EmptyStrMap = &StrMap{0, EmptyHAMT}

// Len returns the number of entries
func (m *StrMap) Len() int { return m.l }

// Get finds value for key. Returns nil if not found.
func (m *StrMap) Get(key string) interface{} {
	v := StrKeyValue{StrValue{StrHash(key), key}, nil}
	v2 := m.m.Lookup(v.H, &v)
	if v2 != nil {
		return v2.(*StrKeyValue).V
	}
	return nil
}

// GetCheck finds value for key and returns a boolean indicating success.
// Useful alternative to Get in case nil values are stored in the map.
func (m *StrMap) GetCheck(key string) (interface{}, bool) {
	v := StrKeyValue{StrValue{StrHash(key), key}, nil}
	v2 := m.m.Lookup(v.H, &v)
	if v2 != nil {
		return v2.(*StrKeyValue).V, true
	}
	return nil, false
}

// Has returns true if key is in m
func (m *StrMap) Has(key string) bool {
	v := StrKeyValue{StrValue{StrHash(key), key}, nil}
	return m.m.Lookup(v.H, &v) != nil
}

// Set returns a StrMap with v
func (m *StrMap) Set(key string, value interface{}) *StrMap {
	v := &StrKeyValue{StrValue{StrHash(key), key}, value}
	len2 := m.l + 1
	m2 := m.m.Insert(0, v.H, v, &len2)
	return &StrMap{len2, m2}
}

// Del returns a StrMap without v. If v is not found, returns the receiver.
func (m *StrMap) Del(key string) *StrMap {
	v := StrKeyValue{StrValue{StrHash(key), key}, nil}
	m2 := m.m.Remove(v.H, &v)
	if m2 == m.m {
		return m // not found; no change
	}
	if m.l == 1 {
		return EmptyStrMap
	}
	return &StrMap{m.l - 1, m2}
}

// Range iterates over all entries by calling f(k,v). If f returns false, iteration stops.
func (m *StrMap) Range(f func(key string, value interface{}) bool) {
	m.m.Range(func(v Value) bool {
		kv := v.(*StrKeyValue)
		return f(kv.K, kv.V)
	})
}

// String returns human-readable text in the format {"key": value, ...}
func (m *StrMap) String() string {
	var sb strings.Builder
	sep := ", "
	if m.l > 5 {
		sep = ",\n  "
		sb.WriteString("{\n  ")
	} else {
		sb.WriteByte('{')
	}
	first := true
	m.Range(func(key string, value interface{}) bool {
		if first {
			first = false
		} else {
			sb.WriteString(sep)
		}
		fmt.Fprintf(&sb, "%q: %v", key, value)
		return true
	})
	if m.l > 5 {
		sb.WriteString(",\n}")
	} else {
		sb.WriteByte('}')
	}
	return sb.String()
}

// GoString returns a Go value representation in the format {{key, value}, ...}
func (m *StrMap) GoString() string {
	var sb strings.Builder
	sb.WriteByte('{')
	first := true
	m.Range(func(key string, value interface{}) bool {
		if first {
			first = false
		} else {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "{%#v, %#v}", key, value)
		return true
	})
	sb.WriteByte('}')
	return sb.String()
}
