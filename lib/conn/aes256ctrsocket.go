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
	. "github.com/stealpipe/lib/encrypt/ctr"
	. "github.com/stealpipe/lib/debug"
)

type Aes256CTRSocket struct {

	conn              Pipe
	key               []byte
	riv               []byte
	wiv               []byte
	speed             int64
	fakeHeaderLength  int64
	fakepackageLength int64
	wctr              cipher.Stream
	rctr              cipher.Stream
	mark              []byte
}

func (p *Aes256CTRSocket) Init(n Pipe, k []byte) {
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

func (p *Aes256CTRSocket) CreateIv() bool {

	if _, err := io.ReadFull(crand.Reader, p.wiv); err != nil {
		panic(err)
	}
	return true
}

func (p *Aes256CTRSocket) ReadFakeHead() bool {

	Debug("ReadFakeHead", p, p.fakeHeaderLength)
	buff := make([]byte, p.fakeHeaderLength)
	b := ReadByte(p.conn, int(p.fakeHeaderLength), buff)
	if !b {
		return b
	}
	Debug("read fake header ", buff)
	lastpath := int(buff[p.fakeHeaderLength-1])
	Debug("ReadFakeHead patch ", lastpath)
	b = ReadByte(p.conn, lastpath, nil)

	return b
}

func (p *Aes256CTRSocket) ReadPackageFake() bool {

	return ReadByte(p.conn, int(p.fakepackageLength), nil)
}

func (p *Aes256CTRSocket) GetConn() net.Conn {
	return p.conn
}

func (p *Aes256CTRSocket) ReadIv() bool {

	Debug("Read Iv", p)

	return ReadByte(p.conn, aes.BlockSize, p.riv)
}

func (p *Aes256CTRSocket) ReadyRead() bool {
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
	block, err := aes.NewCipher(p.key)
	if err != nil {
		panic(err)
	}

	p.rctr = cipher.NewCTR(block, p.riv)

	b, buf := p.Readn(len(Shakehandkey))
	if !b {
		return false
	}
	Debug("read shakehand", string(buf))
	if string(buf) != Shakehandkey {
		return false
	}

	b, buf = p.Readn(4)
	if !b {
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

func (p *Aes256CTRSocket) ReadyWrite() bool {

	sr := rand.New(rand.NewSource(time.Now().UnixNano()))
	patchl := sr.Intn(255)

	patchl = 5

	totalsize := 0
	timestamplen := 4

	r := p.CreateIv()
	if !r {
		return r
	}
	block, err := aes.NewCipher(p.key)
	if err != nil {
		panic(err)
	}

	p.wctr = cipher.NewCTR(block, p.wiv)

	totalsize = int(p.fakeHeaderLength) + patchl + len(p.wiv) + len(Shakehandkey) + timestamplen
	Debug("total size", totalsize, p.fakeHeaderLength, patchl, len(p.wiv), len(Shakehandkey), timestamplen)
	header := make([]byte, totalsize)
	index := int(p.fakeHeaderLength) + patchl
	if _, err := io.ReadFull(crand.Reader, header[0:index]); err != nil {
		panic(err)
	}
	header[p.fakeHeaderLength-1] = byte(patchl)

	Debug("fakeheader ", header)

	Debug("Send iv", p.wiv)
	copy(header[index:], p.wiv)
	index += len(p.wiv)

	Debug("Send shakehand", Shakehandkey)

	_, encryptedbuf, _ := Aes256CTREncrypt([]byte(Shakehandkey), len(Shakehandkey), p.wctr)

	copy(header[index:], encryptedbuf)

	index += len(Shakehandkey)

	timeNow := uint32(time.Now().UTC().Unix())
	btime := make([]byte, timestamplen)
	binary.BigEndian.PutUint32(btime, timeNow)
	Debug("ReadyWrite client time", timeNow)
	_, encryptedbuf, _ = Aes256CTREncrypt(btime, len(btime), p.wctr)

	copy(header[index:], encryptedbuf)
	index += timestamplen
	if index != totalsize {
		panic(index)
	}
	Debug("send fakeheader ", header)
	_, err = SendData(p.conn, header)
	if err != nil {
		return false
	}
	return true
}

func (p *Aes256CTRSocket) Close() {
	p.conn.Close()
}

func (p *Aes256CTRSocket) SetSpeed(speed int64) {
	p.speed = speed
}

func (p *Aes256CTRSocket) SetFakeHeaderLength(len int64) {
	len = 10
	p.fakeHeaderLength = len
	return
	if len < 1500 {
		len = 1951
	}
	if len >= 5000 {
		len = 5212
	}

}

func (p *Aes256CTRSocket) SetFakeHeaderPaddingIndex(int64) {

}

func (p *Aes256CTRSocket) SetPackageFakeLength(int64) {
	p.fakepackageLength = 0
}

func (p *Aes256CTRSocket) Read() (bool, []byte) {

	readbuf := make([]byte, 1024)
	readedlen, err := p.conn.Read(readbuf)
	if err != nil {
		return false, nil
	}
	_, decryptedbuf, _ := Aes256CTRDecrypt(readbuf, readedlen, p.rctr)
	return true, decryptedbuf

}

func (p *Aes256CTRSocket) Readn(l int) (bool, []byte) {

	readbuf := make([]byte, l)
	b := ReadByte(p.conn, l, readbuf)
	if !b {
		return false, nil
	}
	_, decryptedbuf, _ := Aes256CTRDecrypt(readbuf, l, p.rctr)
	return true, decryptedbuf
}

func (p *Aes256CTRSocket) Write(data []byte, datalen int) bool {
	_, encryptedbuf, _ := Aes256CTREncrypt(data, datalen, p.wctr)

	_, err := SendData(p.conn, encryptedbuf)
	if err != nil {
		return false
	}
	return true
}

func (p *Aes256CTRSocket) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *Aes256CTRSocket) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}
