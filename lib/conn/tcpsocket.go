package conn

import (
	"net"
)

const (
	TCPBUFF = 1024 * 64
)

type TCPSocket struct {
	conn Pipe
}

func (p *TCPSocket) Init(n Pipe, k []byte) {
	p.conn = n

}

func (p *TCPSocket) CreateIv() bool {

	return true
}

func (p *TCPSocket) ReadFakeHead() bool {

	return true
}

func (p *TCPSocket) ReadPackageFake() bool {

	return true
}

func (p *TCPSocket) GetConn() net.Conn {
	return p.conn
}

func (p *TCPSocket) ReadIv() bool {

	return true
}

func (p *TCPSocket) ReadyRead() bool {

	return true

}

func (p *TCPSocket) ReadyWrite() bool {

	return true
}

func (p *TCPSocket) Close() {
	p.conn.Close()
}

func (p *TCPSocket) SetSpeed(speed int64) {

}

func (p *TCPSocket) SetFakeHeaderLength(len int64) {

}

func (p *TCPSocket) SetPackageFakeLength(int64) {

}
func (p *TCPSocket) SetFakeHeaderPaddingIndex(int64) {

}
func (p *TCPSocket) Read() (bool, []byte) {

	readbuf := make([]byte, TCPBUFF)
	l, err := p.conn.Read(readbuf)
	if err != nil {
		return false, nil
	}

	return true, readbuf[0:l]
}

func (p *TCPSocket) Readn(l int) (bool, []byte) {
	readnbuf := make([]byte, l)
	b := ReadByte(p.conn, int(l), readnbuf)
	if !b {
		return b, nil
	}
	return true, readnbuf
}

func (p *TCPSocket) Write(data []byte, datalen int) bool {
	_, err := SendData(p.conn, data[0:datalen])
	if err != nil {
		return false
	}
	return true
}

func (p *TCPSocket) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *TCPSocket) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}
