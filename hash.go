// +build !amd64,!aarch64

package immutable

// StrHash returns a non-cryptographic hash for string s, to be used for keys in a trie.
// It's implemented as FNV1a.
func StrHash(s string) uint {
	prime := 0x01000193 // pow(2,24) + pow(2,8) + 0x93
	hash := 0x811C9DC5
	for i := 0; i < len(s); i++ {
		hash = (uint(s[i]) ^ hash) * prime
	}
	return hash
}
