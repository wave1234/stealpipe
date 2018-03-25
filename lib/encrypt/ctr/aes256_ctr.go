package ctr

import (
	"crypto/cipher"
)

func Aes256CTREncrypt(originalData []byte, originalDataLen int, ctr cipher.Stream) (bool, []byte, int) {

	b := make([]byte, originalDataLen)
	ctr.XORKeyStream(b, originalData[0:originalDataLen])
	return true, b, originalDataLen

}

func Aes256CTRDecrypt(encryptedBuf []byte, encryptedLen int, ctr cipher.Stream) (bool, []byte, int) {

	b := make([]byte, encryptedLen)
	ctr.XORKeyStream(b, encryptedBuf[0:encryptedLen])
	return true, b, encryptedLen

}
