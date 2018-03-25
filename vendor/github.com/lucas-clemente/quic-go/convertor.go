package quic

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"math/rand"
	"time"
)

const (
	FakeHeader = aes.BlockSize
)

type Convertor struct {
	UdpEncode bool
	AESIv     [aes.BlockSize]byte
	AESKey    [aes.BlockSize]byte
}

func (c *Convertor) Init(config *Config) {
	c.AESIv = config.AESIv
	c.AESKey = config.AESKey
	c.UdpEncode = config.UdpEncode
	rand.Seed(time.Now().UnixNano())
}

func (c *Convertor) MixData(p []byte) []byte {
	appended := FakeHeader * 2
	if len(p) < 1400 {
		appended += rand.Intn(1400 - len(p) + 1)
	} else {
		appended += rand.Intn(5)
	}
	b := GetEncryptBlockMode(c.AESIv[:], c.AESKey[:])
	_, p2, _ := Aes256CBCEncrypt(p, len(p), b, appended+len(p))
	return p2

}

func (c *Convertor) UnMixData(p []byte, l1 int) (l int) {

	l = l1
	if l == 0 || (l/aes.BlockSize <= 1) {
		return l
	}

	b := GetDecryptBlockMode(c.AESIv[:], c.AESKey[:])
	_, _, l2 := Aes256CBCDecrypt(p, l, b)
	return l2
}

func (c *Convertor) Encoder(p []byte) (p2 []byte) {

	return c.MixData(p)
}

func (c *Convertor) Decoder(p []byte, l int) (l2 int) {
	return c.UnMixData(p, l)
}

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

func Aes256CBCEncrypt(originalData []byte, originalDataLen int, blockMode cipher.BlockMode, encryptedLen int) (bool, []byte, int) {

	l := (FakeHeader + originalDataLen) % aes.BlockSize
	if l == 0 {
		l = (FakeHeader + originalDataLen) / aes.BlockSize * aes.BlockSize
	} else {
		l = (FakeHeader+originalDataLen)/aes.BlockSize*aes.BlockSize + aes.BlockSize
	}

	if l > encryptedLen {
		panic(l)
	}
	data2 := make([]byte, encryptedLen)
	data1 := make([]byte, l)

	rand.Read(data1[0:FakeHeader])
	rand.Read(data2[FakeHeader:])
	binary.BigEndian.PutUint16(data1[2:], uint16(originalDataLen))
	data1[0] = 'M'
	data1[1] = 'K'
	copy(data1[FakeHeader:], originalData)
	blockMode.CryptBlocks(data2, data1)
	return true, data2, encryptedLen
}

func Aes256CBCDecrypt(encryptedBuf []byte, encryptedLen int, blockMode cipher.BlockMode) (bool, []byte, int) {

	l := encryptedLen / aes.BlockSize * aes.BlockSize
	if l < FakeHeader+aes.BlockSize {
		return false, nil, encryptedLen
	}
	blockMode.CryptBlocks(encryptedBuf, encryptedBuf[0:l])

	if encryptedBuf[0] != 'M' && encryptedBuf[1] != 'K' {
		return false, nil, l
	}
	newdatalen := int(binary.BigEndian.Uint16(encryptedBuf[2:]))
	if newdatalen > l+FakeHeader*2 {
		return false, encryptedBuf, encryptedLen
	}
	copy(encryptedBuf, encryptedBuf[FakeHeader:FakeHeader+newdatalen])
	return true, encryptedBuf, newdatalen
}
