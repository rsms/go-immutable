// +build amd64 aarch64

package immutable

// StrHash returns a non-cryptographic hash for string s, to be used for keys in a trie.
// It's implemented as FNV1a.
func StrHash(s string) uint {
	prime := uint(0x100000001B3) // pow(2,40) + pow(2,8) + 0xb3
	hash := uint(0xCBF29CE484222325)
	for i := 0; i < len(s); i++ {
		hash = (uint(s[i]) ^ hash) * prime
	}
	return hash
}
