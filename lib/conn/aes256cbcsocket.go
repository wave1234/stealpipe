package conn

import (
	"crypto/aes"
	"encoding/binary"
	crand "crypto/rand"
	"math/rand"
	"net"
	"crypto/cipher"
	"io"
	"time"
	. "github.com/stealpipe/lib/encrypt/cbc"
	. "github.com/stealpipe/lib/debug"
)

const (
	MaskLen = 18
)

type Aes256CBCSocket struct {
	conn              Pipe
	key               []byte
	riv               []byte
	wiv               []byte
	speed             int64
	fakeHeaderLength  int64
	fakeHeaderIndex   int64
	fakepackageLength int64
	mark              []byte
	rblock            cipher.BlockMode
	wblock            cipher.BlockMode
}

func (p *Aes256CBCSocket) Init(n Pipe, k []byte) {
	p.conn = n
	p.key = k
	p.riv = make([]byte, aes.BlockSize)
	p.wiv = make([]byte, aes.BlockSize)

	if _, err := io.ReadFull(crand.Reader, p.riv); err != nil {
		panic(err)
	}

	if _, err := io.ReadFull(crand.Reader, p.wiv); err != nil {
		panic(err)
	}
	p.mark = make([]byte, len(Shakehandkey))
}

func (p *Aes256CBCSocket) CreateIv() bool {
	if _, err := io.ReadFull(crand.Reader, p.wiv); err != nil {
		panic(err)
	}
	return true
}

func (p *Aes256CBCSocket) ReadFakeHead() bool {

	Debug("ReadFakeHead", p, p.fakeHeaderLength)
	buff := make([]byte, p.fakeHeaderLength)
	b := ReadByte(p.conn, int(p.fakeHeaderLength), buff)
	if !b {
		return b
	}
	lastpath := int(buff[p.fakeHeaderIndex])

	b = ReadByte(p.conn, lastpath, nil)

	return b
}

func (p *Aes256CBCSocket) ReadPackageFake() bool {

	return ReadByte(p.conn, int(p.fakepackageLength), nil)
}

func (p *Aes256CBCSocket) GetConn() net.Conn {
	return p.conn
}

func (p *Aes256CBCSocket) ReadIv() bool {

	Debug("Read Iv", p)
	b := ReadByte(p.conn, aes.BlockSize, p.riv)
	Debug("Read Iv", p.riv)
	return b
}

func (p *Aes256CBCSocket) ReadyRead() bool {
	Debug("ReadyRead Read fake head")
	r := p.ReadFakeHead()
	if !r {
		return r
	}

	Debug("read iv")
	r = p.ReadIv()
	if !r {
		return r
	}

	Debug("riv", p.riv)
	p.rblock = GetDecryptBlockMode(p.key, p.riv)

	b, buf := p.Read()

	if !b {
		return false
	}
	if len(buf) != len(Shakehandkey) {
		Debug("len(buf) != len(Shakehandkey)", len(buf), len(Shakehandkey))
		return false
	}
	Debug("read shakehand", string(buf))
	if string(buf) != Shakehandkey {
		return false
	}

	b, buf = p.Read()
	if !b {
		return false
	}

	if len(buf) != 4 {
		Debug("len(buf) != 4", buf)
		return false
	}
	clientTimeNow := binary.BigEndian.Uint32(buf)

	Debug("ReadyRead client time", clientTimeNow)
	timeNow := uint32(time.Now().UTC().Unix())
	if (clientTimeNow < timeNow-PACKTTL) || (clientTimeNow > timeNow+PACKTTL) {
		return false
	}
	return true

}

func (p *Aes256CBCSocket) ReadyWrite() bool {

	sr := rand.New(rand.NewSource(time.Now().UnixNano()))

	fackeHeaderPaddingLength := sr.Intn(255)

	fakebyte := make([]byte, p.fakeHeaderLength+int64(fackeHeaderPaddingLength))
	if _, err := io.ReadFull(crand.Reader, fakebyte); err != nil {
		panic(err)
	}
	fakebyte[p.fakeHeaderIndex] = byte(fackeHeaderPaddingLength)

	Debug("Send fakeHead", p.fakeHeaderLength+int64(fackeHeaderPaddingLength))
	_, err := SendData(p.conn, fakebyte)
	if err != nil {
		return false
	}
	p.CreateIv()
	p.wblock = GetEncryptBlockMode(p.key, p.wiv)
	_, err = SendData(p.conn, p.wiv)
	if err != nil {
		return false
	}
	Debug("ReadyWrite write iv", p.wiv)
	b := p.Write([]byte(Shakehandkey), len(Shakehandkey))
	if !b {
		return b
	}

	timeNow := uint32(time.Now().UTC().Unix())
	btime := make([]byte, 4)
	binary.BigEndian.PutUint32(btime, timeNow)

	Debug("ReadyWrite client time", timeNow)
	b = p.Write(btime, 4)

	return b
}

func (p *Aes256CBCSocket) Close() {
	p.conn.Close()
}

func (p *Aes256CBCSocket) SetSpeed(speed int64) {
	p.speed = speed
}

func (p *Aes256CBCSocket) SetFakeHeaderLength(len int64) {
	if len < 1300 {
		len = 1500
	}
	if len >= 5000 {
		len = 5000
	}
	p.fakeHeaderLength = len
}
func (p *Aes256CBCSocket) SetFakeHeaderPaddingIndex(len int64) {

	if p.fakeHeaderLength > len {
		p.fakeHeaderIndex = p.fakeHeaderLength - len
	} else {
		p.fakeHeaderIndex = p.fakeHeaderLength - 1
	}

}

func (p *Aes256CBCSocket) SetPackageFakeLength(int64) {
	p.fakepackageLength = 0
}

func (p *Aes256CBCSocket) Read() (bool, []byte) {

	// read header
	readbuf := make([]byte, 2*aes.BlockSize)

	b := ReadByte(p.conn, 2*aes.BlockSize, readbuf)
	if !b {
		//panic(false)
		return b, nil
	}
	Debug("read enHead", readbuf)
	b, headbuf, l2 := Aes256CBCDecrypt(readbuf, 2*aes.BlockSize, p.rblock)
	if !b {
		//panic(false)
		return b, nil
	}
	Debug("read xx header", headbuf, l2)
	datalen := binary.BigEndian.Uint16(headbuf)

	patchl := binary.BigEndian.Uint16(headbuf[2:])
	mask := headbuf[4 : MaskLen+4]
	Debug(headbuf)
	Debug(MaskLen)
	Debug(mask)
	Debug(string(mask))
	Debug(Shakehandkey)
	if string(mask) != Shakehandkey[0:MaskLen] {
		return false, nil
	}

	// read header
	readbuf = make([]byte, datalen)
	b = ReadByte(p.conn, int(datalen), readbuf)
	if !b {
		return b, nil
	}
	if patchl > 0 {
		b = ReadByte(p.conn, int(patchl), nil)
	}
	_, decryptedbuf, _ := Aes256CBCDecrypt(readbuf, int(datalen), p.rblock)
	return true, decryptedbuf
}

func (p *Aes256CBCSocket) Readn(l int) (bool, []byte) {
	readnbuf := make([]byte, l)
	index := 0

	for {
		b, buf := p.Read()
		if !b {
			return b, nil
		}
		if len(buf)+index > l {
			panic("wrong data package")
		}
		for i := 0; i < len(buf); i++ {
			readnbuf[index+i] = buf[i]
		}
		index += len(buf)
		if index == l {
			break
		}
	}
	return true, readnbuf
}

func (p *Aes256CBCSocket) Write(data []byte, datalen int) bool {
	begin := 0
	endindex := 0
	sr := rand.New(rand.NewSource(time.Now().UnixNano()))
	patchl := 0
	for datalen > begin {
		if datalen-begin > MAXPACKLENGTH {
			endindex = begin + MAXPACKLENGTH
		} else {
			endindex = datalen
		}

		sendDatalen := endindex - begin
		Debug("sendDatalen : ", sendDatalen)
		header := make([]byte, 2*aes.BlockSize)
		for i := 0; i < MaskLen; i++ {
			header[4+i] = Shakehandkey[i]
		}

		l4 := Pkcs5Padd(header, MaskLen+2+4)
		l3 := Pkcs5Padd(data[begin:endindex], sendDatalen)

		if l3+l4 < 1300 {
			patchl = sr.Intn(1300 - (l3 + l4))
		}

		binary.BigEndian.PutUint16(header, uint16(l3))
		binary.BigEndian.PutUint16(header[2:], uint16(patchl))

		Debug("Write head ", sendDatalen, l3, header)
		_, encryptedbuf, l2 := Aes256CBCEncrypt(header, MaskLen+4, p.wblock)
		Debug("Write head Enc header ", sendDatalen, aes.BlockSize, "l4", l4, " l2: ", l2, "l3 ", l3, "encryptedbuf : ", len(encryptedbuf))
		Debug("Write  enhead ", encryptedbuf)

		_, encryptedbuf2, _ := Aes256CBCEncrypt(data[begin:endindex], sendDatalen, p.wblock)
		Debug("Write data ", l3, encryptedbuf2)

		SendBuffer := append(encryptedbuf, encryptedbuf2...)
		if patchl > 0 {
			radombuf := make([]byte, patchl)
			if _, err := io.ReadFull(crand.Reader, radombuf); err != nil {
				panic(err)
			}
			SendBuffer = append(SendBuffer, radombuf...)
		}

		_, err := SendData(p.conn, SendBuffer)
		if err != nil {
			return false
		}

		begin += sendDatalen
	}

	return true
}

func (p *Aes256CBCSocket) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *Aes256CBCSocket) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}
