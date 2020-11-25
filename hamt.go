package immutable

import (
	"fmt"
	"strings"

	"github.com/rsms/go-bits"
)

const intSize uint = 32 << (^uint(0) >> 63) // bits of int on target platform
const uintMax uint = ^uint(0)

const hamtBranches = uint(intSize)             // uint8(32)
const hamtBits = uint(7 - (64 / hamtBranches)) // 32=5, 64=6
const hamtMask = uint(hamtBranches) - 1        // 32=...0011111, 64=...0111111 (5 vs 6 bits set)

var EmptyHAMT = &HAMT{}

// HAMT implements an immutable persistent Hash Array Mapped Trie
type HAMT struct {
	bmap    uint          // bitmap; router for entries
	entries []interface{} // Value | *HAMT | *hcollision
}

// hcollision houses values with identical keys
type hcollision []Value

// Empty returns true if the HAMT does not contain any entries
func (m *HAMT) Empty() bool {
	return m == nil || len(m.entries) == 0
}

// Lookup retrieves the value for an entry identified by key+v
func (m *HAMT) Lookup(key uint, v Value) Value {
	shift := uint(0)
	for {
		// See mutInsert() for detail description of the algorithm.
		// Check if index bit is set in bitmap
		bitpos := uint(1) << ((key >> shift) & hamtMask)
		if m.bmap&bitpos == 0 {
			return nil
		}
		// Compare to value at m.entries[bi]
		// where bi is the bucket index by mapping index bit -> bucket index.
		switch e := m.entries[bits.Bitindex(m.bmap, bitpos)].(type) {
		case *HAMT:
			m = e
		case hcollision:
			for _, v1 := range e {
				if v1.Equal(v) {
					return v1
				}
			}
			return nil
		case Value:
			if e.Equal(v) {
				return e
			}
			return nil
		}
		shift += hamtBits
	}
}

// Insert returns a new HAMT with value v.
// resized is decremented by 1 in case the operation replaced an existing entry.
func (m *HAMT) Insert(shift, key uint, v Value, resized *int) *HAMT {
	bitpos := uint(1) << ((key >> shift) & hamtMask) // key bit position
	bi := bits.Bitindex(m.bmap, bitpos)              // bucket index
	// Now, one of three cases may be encountered:
	//
	// 1. The entry is empty indicating that the key is not in the tree.
	//    The value is inserted into the HAMT.
	//
	// 2. The entry is a Value.
	//    If a lookup is performed, check for a match to determine success or failure.
	//    If an insertion is performed, one of the following cases are encountered:
	//
	//    2.1. The existing value v1 is equivalent to the new value v2; v1 is replaced by v2.
	//
	//    2.2. Existing v1 is different from v2 but shares the same key (i.e. hash collision.)
	//         v1 and v2 are moved into a hcollision list; v1 is replaced by this hcollision list.
	//
	//    2.3. v1 and v2 are different with different keys. A HAMT m2 is created with v1 and v2
	//         at entries2[0,2]={v1,v2}. v1 is replaced by the HAMT m2.
	//
	// 3. The entry is a HAMT; called a sub-hash table in the original HAMT paper.
	//    Evalute steps 1-3 on the map by calling mutInsert() recursively.
	//
	// 4. The entry is a hcollision list.
	//
	if m.bmap&bitpos == 0 {
		// empty; index bit not set in bmap. Set the bit and append value to entries list.
		// copy entries in m2 with +1 space for slot at bi
		m2 := &HAMT{m.bmap | bitpos, make([]interface{}, len(m.entries)+1)}
		copy(m2.entries, m.entries[:bi])
		copy(m2.entries[bi+1:], m.entries[bi:])
		m2.entries[bi] = v
		return m2
	}

	// an entry or branch occupies the slot.
	// Note: Consider converting this function to use iteration instead of recursion.
	//       If/when doing so, benchmark to make sure it is actually more efficient.
	//       (With interface{} in Go, it often is but not always.)
	m2 := &HAMT{m.bmap, make([]interface{}, len(m.entries))}
	copy(m2.entries, m.entries)
	switch e := m.entries[bi].(type) {
	case *HAMT:
		// enter branch
		m2.entries[bi] = e.Insert(shift+hamtBits, key, v, resized)
	case hcollision:
		// existing collision (invariant: last branch; shift >= (hamtBranches-shift))
		m2.entries[bi] = e.withValue(v, resized)
	case Value:
		// A value already exists at this path
		key1 := e.Hash()
		if key1 == key && e.Equal(v) {
			// replace
			(*resized)--
			m2.entries[bi] = v
		} else {
			m2.entries[bi] = makeHamtBranch(shift+hamtBits, key1, key, e, v)
		}
	}
	return m2
}

// Remove deletes an entry identified by key+v
func (m *HAMT) Remove(key uint, v Value) *HAMT {
	var hasCollision bool // temporary state
	return m.remove(0, key, v, &hasCollision)
}

func (m *HAMT) remove(shift, key uint, v2 Value, hasCollision *bool) *HAMT {
	bitpos := uint(1) << ((key >> shift) & hamtMask) // key bit position
	if m.bmap&bitpos != 0 {
		bi := bits.Bitindex(m.bmap, bitpos)
		switch e := m.entries[bi].(type) {
		case *HAMT:
			// enter branch, calling remove() recursively, then either collapse the path into just
			// a value in case remove() returned a HAMT with a single Value, or just copy m with
			// the map returned from remove() at bi.
			//
			// Note: consider making this iterative; non-recursive.
			m3 := e.remove(shift+hamtBits, key, v2, hasCollision)
			if m3 != e {
				m2 := &HAMT{m.bmap, make([]interface{}, len(m.entries))} // copy
				copy(m2.entries, m.entries)
				if len(m3.entries) == 1 && !*hasCollision {
					if _, ismap := m3.entries[0].(*HAMT); !ismap {
						// collapse path
						m2.entries[bi] = m3.entries[0]
						return m2
					}
				}
				m2.entries[bi] = m3
				return m2
			}

		case hcollision:
			if c2, found := e.withoutValue(v2); found {
				m2 := &HAMT{m.bmap, make([]interface{}, len(m.entries))} // copy
				copy(m2.entries, m.entries)
				if len(c2) == 1 {
					// collapse collision
					m2.entries[bi] = c2[0]
				} else {
					*hasCollision = true
					m2.entries[bi] = c2
				}
				return m2
			}

		case Value:
			// m2 = m1 without v2
			if e.Equal(v2) {
				z := len(m.entries)
				if z == 1 {
					return EmptyHAMT
				}
				m2 := &HAMT{m.bmap &^ bitpos, make([]interface{}, z-1)}
				copy(m2.entries[:bi], m.entries[:bi])   // [0..bi)
				copy(m2.entries[bi:], m.entries[bi+1:]) // [bi..END)
				return m2
			}
		}
	}

	return m
}

// makeHamtBranch creates a HAMT at level with two entries v1 and v2.
// In case v1 and v2 are equivalent, this function instead just returns v2 to
// be replaced (does not create a map.)
//
// Returns the entry that represents the new branch, and secondly a boolean value
// indicating if v2 was added (false means an existing value was replaced.)
//
func makeHamtBranch(shift, key1, key2 uint, v1, v2 Value) interface{} {
	// Compute the "path component" for key1 and key2 for level.
	// shift is the new level for the branch which is being created.
	index1 := (key1 >> shift) & hamtMask
	index2 := (key2 >> shift) & hamtMask

	// loop that creates new branches while key prefixes are shared.
	//
	// head and tail of a chain when there is subindex conflict,
	// representing intermediate branches.
	var mHead, mTail *HAMT
	for index1 == index2 {
		if shift >= hamtBranches {
			c := hcollision{v1, v2}
			if mHead == nil {
				return c
			}
			// We have an existing head we build in the loop above.
			// Add c to its tail and return the head.
			mTail.entries[0] = c
			return mHead
		}

		// append to tail of branch list
		m := &HAMT{uint(1) << index1, []interface{}{nil}}
		if mTail == nil {
			mHead = m
		} else {
			mTail.entries[0] = m
		}
		mTail = m

		shift += hamtBits
		index1 = (key1 >> shift) & hamtMask
		index2 = (key2 >> shift) & hamtMask
	}

	// create map with v1,v2
	bmap := (uint(1) << index1) | (uint(1) << index2)
	var m *HAMT
	if index1 < index2 {
		m = &HAMT{bmap, []interface{}{v1, v2}}
	} else {
		m = &HAMT{bmap, []interface{}{v2, v1}}
	}

	if mHead == nil {
		return m
	}

	// We have an existing head we build in the loop above.
	// Add m to its tail and return the head.
	mTail.entries[0] = m
	return mHead
}

// hcollision.withValue returns a copy of c with v
func (c1 hcollision) withValue(v2 Value, resized *int) hcollision {
	// Check for an equivalent value in the collision list to replace, or else append.
	l := len(c1)
	i, z := 0, l+1
	for ; i < l; i++ {
		if c1[i].Equal(v2) {
			z-- // no need to grow array; will be replacing value at i
			(*resized)--
			break
		}
	}
	c2 := make(hcollision, z)
	copy(c2, c1)
	c2[i] = v2
	return c2
}

func (c1 hcollision) withoutValue(v2 Value) (hcollision, bool) {
	for i := 0; i < len(c1); i++ {
		v1 := c1[i]
		if v1.Equal(v2) {
			c2 := make(hcollision, len(c1)-1)
			copy(c2[:i], c1[:i])   // [0..i)
			copy(c2[i:], c1[i+1:]) // [i..END)
			return c2, true
		}
	}
	return c1, false
}

// Range calls f for every entry in the HAMT. If f returns false iteration stops.
// Returns the return value of f.
func (m *HAMT) Range(f func(Value) bool) bool {
	for i := 0; i < len(m.entries); i++ {
		switch e := m.entries[i].(type) {
		case *HAMT:
			if !e.Range(f) {
				return false
			}
		case hcollision:
			for _, v := range e {
				if !f(v) {
					return false
				}
			}
		case Value:
			if !f(e) {
				return false
			}
		}
	}
	return true
}

// Repr returns a human-readable, printable string representation of the HAMT
func (m *HAMT) Repr() string {
	var sb strings.Builder
	m.repr(&sb, 0, "\n  ")
	return sb.String()
}

func (m *HAMT) repr(sb *strings.Builder, level uint, indent string) {
	fmt.Fprintf(sb, "map (level %d)", level)
	for i, e := range m.entries {
		fmt.Fprintf(sb, "%s#%d => ", indent, i)
		switch e := e.(type) {
		case *HAMT:
			e.repr(sb, level+1, indent+"  ")
		case hcollision:
			sb.WriteString("collision")
			for _, v := range e {
				fmt.Fprintf(sb, "%s  - Value %s", indent, v)
			}
		case Value:
			fmt.Fprintf(sb, "Value %s", e)
		}
	}
}
