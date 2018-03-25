package cbc

import (
	"crypto/aes"
	crand "crypto/rand"
	"fmt"

	. "github.com/stealpipe/lib/debug"
	. "github.com/stealpipe/lib/encrypt"
	"io"
	"math/rand"
	"testing"
	"time"
)

const (
	Iterations = 1
)

func Test_aes(t *testing.T) {
	s2 := rand.NewSource(time.Now().Unix())
	r2 := rand.New(s2)
	key := "aaaaaaaaaaaa"
	fullkey := GetFullKey(key)

	started := time.Now()
	datalen := 0

	for testtime := 0; testtime < 3; testtime++ {
		datalen = r2.Intn(100)
		iv := make([]byte, aes.BlockSize)
		srcbyte := make([]byte, datalen)

		for x := 0; x < 1; x++ {

			if _, err := io.ReadFull(crand.Reader, srcbyte); err != nil {
				panic(err)
			}
			if _, err := io.ReadFull(crand.Reader, iv); err != nil {
				panic(err)
			}

			b1 := GetEncryptBlockMode(fullkey, iv)
			b2 := GetDecryptBlockMode(fullkey, iv)

			fmt.Println("src", srcbyte, datalen)
			_, d1, encryptedlen := Aes256CBCEncrypt(srcbyte, datalen, b1)
			fmt.Println("enc", d1, encryptedlen)
			_, d2, decryptedlen := Aes256CBCDecrypt(d1, encryptedlen, b2)
			fmt.Println("dec", d2, decryptedlen)
			if decryptedlen != datalen {
				t.Log("decryptedlen ", decryptedlen)
				t.Log("srclen ", datalen)
				t.Fail()
			}
			for i := 0; i < datalen; i++ {
				if srcbyte[i] != d2[i] {
					t.Log(srcbyte[i] != d2[i])
					t.Fail()
					break
				} else {
					//fmt.Printf("%c", decryptedbuf[i])
				}
			}
			Debug(" encryptedlen ", encryptedlen, " decryptedlen", decryptedlen, " datalen", datalen)
		}

	}
	finished := time.Now()
	Debug(Iterations, finished.Sub(started))
}

func entry(srcbyte []byte, t *testing.T) {
	key := "aaaaaaaaaaaa"
	fullkey := GetFullKey(key)

	datalen := len(srcbyte)
	iv := make([]byte, aes.BlockSize)

	b1 := GetEncryptBlockMode(fullkey, iv)
	b2 := GetDecryptBlockMode(fullkey, iv)

	fmt.Println("src", srcbyte, datalen)
	_, d1, encryptedlen := Aes256CBCEncrypt(srcbyte, datalen, b1)
	fmt.Println("enc", d1, encryptedlen)
	_, d2, decryptedlen := Aes256CBCDecrypt(d1, encryptedlen, b2)
	fmt.Println("dec", d2, decryptedlen)
	if decryptedlen != datalen {
		t.Log("decryptedlen ", decryptedlen)
		t.Log("srclen ", datalen)
		t.Fail()
	}
	for i := 0; i < datalen; i++ {
		if srcbyte[i] != d2[i] {
			t.Log(srcbyte[i] != d2[i])
			t.Fail()
			break
		} else {
			//fmt.Printf("%c", decryptedbuf[i])
		}
	}
	Debug(" encryptedlen ", encryptedlen, " decryptedlen", decryptedlen, " datalen", datalen)

}
func Test_aes2(t *testing.T) {

	entry([]byte("aaa"), t)
	entry([]byte("aaa1"), t)
	entry([]byte("aaa12"), t)

}
