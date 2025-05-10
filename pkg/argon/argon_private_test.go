package argon

import "testing"

func BenchmarkParseHash(b *testing.B) {
	encodedHash, _ := GenerateFromPassword("correct horse battery staple")

	for b.Loop() {
		_, _, _, _, _, _ = parseHash(encodedHash)
	}
}
