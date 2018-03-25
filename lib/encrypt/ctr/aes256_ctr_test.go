package ctr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	. "github.com/stealpipe/lib/debug"
	. "github.com/stealpipe/lib/encrypt"
	"io"
	mrand "math/rand"
	"testing"
	"time"
)

const (
	Iterations = 1
)

func Test_aes1(t *testing.T) {
	s2 := mrand.NewSource(time.Now().Unix())
	r2 := mrand.New(s2)

	key := "aaaaaaaaaaaa"
	fullkey := GetFullKey(key)
	Trace(fullkey)
	started := time.Now()
	datalen := 30

	block, err := aes.NewCipher(fullkey)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize)
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	stream2 := cipher.NewCTR(block, iv)

	for testtime := 0; testtime < 3; testtime++ {
		datalen = r2.Intn(100)
		srcbyte := make([]byte, datalen)

		if _, err := io.ReadFull(rand.Reader, srcbyte); err != nil {
			panic(err)
		}

		for x := 0; x < 1; x++ {

			fmt.Println("src:", srcbyte)
			_, d1, encryptedlen := Aes256CTREncrypt(srcbyte, datalen, stream)
			fmt.Println("enc:", d1)
			_, d2, decryptedlen := Aes256CTRDecrypt(d1, encryptedlen, stream2)
			fmt.Println("dec:", d2)
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
	block, _ := aes.NewCipher(fullkey)

	b1 := cipher.NewCTR(block, iv)
	b2 := cipher.NewCTR(block, iv)

	fmt.Println("src", srcbyte, datalen)
	_, d1, encryptedlen := Aes256CTREncrypt(srcbyte, datalen, b1)
	fmt.Println("enc", d1, encryptedlen)
	_, d2, decryptedlen := Aes256CTREncrypt(d1, encryptedlen, b2)
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
