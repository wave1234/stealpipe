package conn

import (
	"net"
	"time"
)

type Pipe interface {
	Write(sendData []byte) (int, error)
	Read(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

const (
	Shakehandkey  = "sbamtbctatfajtb!!!"
	PACKTTL       = 60 * 5
	MAXPACKLENGTH = 60 * 1000
)

type ConnInterface interface {
	Init(n Pipe, k []byte)
	Read() (bool, []byte)
	Readn(int) (bool, []byte)
	Write([]byte, int) bool
	Close()
	SetSpeed(int64)
	SetFakeHeaderLength(int64)
	SetFakeHeaderPaddingIndex(int64)
	SetPackageFakeLength(int64)
	ReadIv() bool
	CreateIv() bool
	ReadFakeHead() bool
	ReadPackageFake() bool
	ReadyRead() bool
	ReadyWrite() bool
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}
