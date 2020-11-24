package immutable

// strHash returns a non-cryptographic hash for string s, to be used for keys in a trie.
// It's implemented as FNV1a.
func strHash(s string) uint {
	prime := strHashPrime
	hash := strHashInit
	for i := 0; i < len(s); i++ {
		hash = (uint(s[i]) ^ hash) * prime
	}
	return hash
}

var strHashPrime uint = 0x01000193 // pow(2,24) + pow(2,8) + 0x93
var strHashInit uint = 0x811C9DC5

func init() {
	if intSize >= 64 {
		strHashPrime = 0x100000001B3 // pow(2,40) + pow(2,8) + 0xb3
		strHashInit = 0xCBF29CE484222325
	}
}
