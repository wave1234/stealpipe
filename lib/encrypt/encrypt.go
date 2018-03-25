package encrypt

import (
	"crypto/sha1"
	"io"
)

const KEYLENGTH = 32 //256bit

func GetFullKey(password string) []byte {
	h := sha1.New()
	io.WriteString(h, password)
	return h.Sum(nil)[0:KEYLENGTH]
}
