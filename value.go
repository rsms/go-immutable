package immutable

import "fmt"

// Value defines the operations that must be implemented for the value type of a HAMT
type Value interface {
	// Hash should return a "as unique as possible" integer for the "key" of the value
	Hash() uint
	// Equal should return true if the other value's key is equivalent to the receiver
	Equal(Value) bool
}

// StrValue is a type of Value with a string value
type StrValue struct {
	H uint
	K string
}

func NewStrValue(s string) *StrValue {
	return &StrValue{StrHash(s), s}
}

func (e *StrValue) Value() string { return e.K }
func (e *StrValue) SetValue(s string) {
	e.H = StrHash(s)
	e.K = s
}

func (e *StrValue) Hash() uint { return e.H }
func (e *StrValue) Equal(b Value) bool {
	v2 := b.(*StrValue)
	return e.H == v2.H && e.K == v2.K
}
func (e *StrValue) String() string { return e.K }

// StrKeyValue is a type of Value with a string key and interface value
type StrKeyValue struct {
	StrValue
	V interface{}
}

func NewStrKeyValue(key string, value interface{}) *StrKeyValue {
	return &StrKeyValue{StrValue{StrHash(key), key}, value}
}

func (e *StrKeyValue) Value() interface{}     { return e.V }
func (e *StrKeyValue) SetValue(v interface{}) { e.V = v }

func (e *StrKeyValue) Key() string       { return e.K }
func (e *StrKeyValue) SetKey(key string) { e.StrValue.SetValue(key) }

func (e *StrKeyValue) Equal(b Value) bool {
	v2 := b.(*StrKeyValue)
	return e.H == v2.H && e.K == v2.K
}

func (e *StrKeyValue) String() string {
	return fmt.Sprintf("(%#v = %v)", e.K, e.V)
}
