package conn

import (
	crand "crypto/rand"
	. "github.com/stealpipe/lib/encrypt"
	. "github.com/stealpipe/lib/encrypt/cbc"
	"testing"

	"crypto/aes"
	"io"
	"fmt"
)

func BenchmarkEncry(b *testing.B) {
	Datalen := 1500
	buffer1 := make([]byte, Datalen)
	iv1 := make([]byte, aes.BlockSize)
	iv2 := make([]byte, aes.BlockSize)
	key := "12345"
	fullkey := GetFullKey(key)

	if _, err := io.ReadFull(crand.Reader, iv1); err != nil {
		panic(err)
	}
	copy(iv1, iv2)

	b1 := GetEncryptBlockMode(fullkey, iv1)
	b2 := GetDecryptBlockMode(fullkey, iv2)

	for i := 0; i < 100000; i++ {
		if _, err := io.ReadFull(crand.Reader, buffer1); err != nil {
			panic(err)
		}

		bResult, output, outputLength := Aes256CBCEncrypt(buffer1, Datalen, b1)
		if !bResult {
			panic(bResult)
		}
		if len(output) != outputLength {
			panic(output)
		}
		bResult2, input, inputLentg := Aes256CBCDecrypt(output, outputLength, b2)
		if !bResult2 {
			panic(bResult)
		}
		if len(input) != inputLentg {
			panic(input)
		}
		for i, v := range input {
			if v != buffer1[i] {
				fmt.Println(buffer1, input)
				panic(i)
			}
		}

	}
}
