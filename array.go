package immutable

import (
  "fmt"
  "strings"
)

// Array stores Value objects in a HAMT structure
type Array struct {
  l int   // number of entries
  h *HAMT // trie root
}

type arrayValue struct { // conforms to Value
  index uint
  value interface{}
}

func (e *arrayValue) Hash() uint              { return e.index }
func (e *arrayValue) Equal(otherv Value) bool { return e.index == otherv.(*arrayValue).index }

func (e *arrayValue) String() string { return fmt.Sprint(e.value) }

// The empty Array
var EmptyArray = &Array{0, EmptyHAMT}

// Len returns the number of entries in the array
func (a *Array) Len() int { return a.l }

// Get finds value for v. Returns nil if not found.
func (a *Array) Get(index int) interface{} {
  e := arrayValue{uint(index), nil}
  if e := a.h.Lookup(e.index, &e); e != nil {
    return e.(*arrayValue).value
  }
  return nil
}

// Set a value at index
func (a *Array) Set(index int, value interface{}) *Array {
  uindex := uint(index)
  e := &arrayValue{uindex, value}
  len2 := a.l + 1
  h2 := a.h.Insert(0, uindex, e, &len2)
  return &Array{len2, h2}
}

// Del returns a Array without v. If v is not found, returns the receiver.
func (a *Array) Del(index int) *Array {
  e := arrayValue{uint(index), nil}
  h2 := a.h.Remove(e.index, &e)
  if h2 == a.h {
    return a // not found; no change
  }
  return &Array{a.l - 1, h2}
}

// String returns human-readable text in the format "{Value, Value, Value}"
func (a *Array) String() string { return stringArray(a.h) }

func stringArray(m *HAMT) string {
  var sb strings.Builder
  sb.WriteByte('[')
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
  sb.WriteByte(']')
  return sb.String()
}
