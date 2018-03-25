package cbc

import (
	"crypto/aes"
	"crypto/cipher"
)

func Pkcs5Padd(originalData []byte, originalDataLength int) int {

	paddingLength := aes.BlockSize - originalDataLength%aes.BlockSize
	changed := originalDataLength + paddingLength
	return changed
}

func pkcs5Padding(originalData []byte, originalDataLength int) ([]byte, int) {
	paddingLength := aes.BlockSize - originalDataLength%aes.BlockSize
	changed := originalDataLength + paddingLength
	org := make([]byte, changed)

	for i := 0; i < originalDataLength; i++ {
		org[i] = originalData[i]
	}
	for i := 0; i < paddingLength; i++ {
		org[originalDataLength+i] = byte(paddingLength)
	}
	return org, changed
}

func pkcs5UnPadding(originalData []byte, len int) int {

	if len < int(originalData[len-1]) {
		return -1
	}
	return len - int(originalData[len-1])
}

func GetEncryptBlockMode(key []byte, iv []byte) cipher.BlockMode {

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockMode := cipher.NewCBCEncrypter(block, iv)
	return blockMode
}

func GetDecryptBlockMode(key []byte, iv []byte) cipher.BlockMode {

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	return blockMode
}

func Aes256CBCEncrypt(originalData []byte, originalDataLen int, blockMode cipher.BlockMode) (bool, []byte, int) {

	originalData, paddedLength := pkcs5Padding(originalData, originalDataLen)
	encryptedBufLength := paddedLength
	encryptedBuf := make([]byte, encryptedBufLength)
	blockMode.CryptBlocks(encryptedBuf, originalData)
	return true, encryptedBuf, encryptedBufLength
}

func Aes256CBCDecrypt(encryptedBuf []byte, encryptedLen int, blockMode cipher.BlockMode) (bool, []byte, int) {

	if (encryptedLen%aes.BlockSize != 0) || encryptedLen <= 0 {
		return false, nil, 0
	}
	decryptedBuf := make([]byte, encryptedLen)
	blockMode.CryptBlocks(decryptedBuf, encryptedBuf)
	originalDataLen := pkcs5UnPadding(decryptedBuf, encryptedLen)
	if originalDataLen < 0 {
		return false, nil, 0
	}
	return true, decryptedBuf[:originalDataLen], originalDataLen
}
