package conn

import (
	"errors"
	. "github.com/stealpipe/lib/debug"
	"net"
	"sync"
	"time"
)

type MockConn struct {
	data    []byte
	ch      chan int
	bufflen int
	datalen int
	lock    sync.Mutex
	closed  int
}

type MakePipe struct {
	p1 *MockConn
	p2 *MockConn
}

func (p *MakePipe) Init() {
	p.p1 = &MockConn{}
	p.p1.Init()
	p.p2 = &MockConn{}
	p.p2.Init()
	Debug("make Pipe Init", p, p.p1, p.p2)
}

func (p *MakePipe) GetClient() Pipe {

	px := MockPipe{}
	px.Init(p.p2, p.p1)

	Debug("make Pipe px client", px, px.p1, px.p2)
	return px
}

func (p *MakePipe) GetServer() Pipe {
	px := MockPipe{}
	px.Init(p.p1, p.p2)

	Debug("make Pipe px server", &px, px.p1, px.p2)
	return px
}

type MockPipe struct {
	p1 *MockConn
	p2 *MockConn
}

func (p *MockPipe) Init(p1, p2 *MockConn) {
	Debug("mock pipe", &p, p1, p2)
	p.p1 = p1
	p.p2 = p2
}

func (p MockPipe) Write(sendData []byte) (int, error) {

	Debug("mock pipe write", &p, p.p1)
	if p.p1 == nil {
		panic(p)
	}
	return p.p1.Write(sendData)
}

func (p MockPipe) Read(b []byte) (n int, err error) {

	Debug("mock pipe read", &p, p.p2)
	if p.p2 == nil {
		panic(p)
	}

	return p.p2.Read(b)
}

func (p MockPipe) Close() error {
	p.p1.Close()
	return p.p2.Close()

}

type PipeAddr struct {
}

func (p PipeAddr) Network() string { // name of the network (for example, "tcp", "udp")
	return "pipe"
}

func (p PipeAddr) String() string { // string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
	return "19.49.11.28"
}

func (p MockPipe) LocalAddr() net.Addr {

	return PipeAddr{}
}
func (p MockPipe) RemoteAddr() net.Addr {

	return PipeAddr{}
}
func (p MockPipe) SetDeadline(t time.Time) error {

	return nil
}

func (p MockPipe) SetReadDeadline(t time.Time) error {

	return nil
}

func (p MockPipe) SetWriteDeadline(t time.Time) error {

	return nil

}

func (p *MockConn) Init() {
	bl := 1024 * 1024
	p.data = make([]byte, bl)
	p.ch = make(chan int, 100)
	p.bufflen = bl
}

func (p *MockConn) Write(sendData []byte) (int, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.closed != 0 {
		panic(" write to closed conn")
		return -1, errors.New("closed")
	}
	sendch := false
	if p.datalen == 0 {
		sendch = true
	}
	Info("conn Write ", len(sendData), p.datalen, p.bufflen)
	if p.datalen+len(sendData) < p.bufflen {
	} else {
		p.bufflen = p.datalen + len(sendData)
		newbuf := make([]byte, p.bufflen)
		for i := 0; i < p.datalen; i++ {
			newbuf[i] = p.data[i]
		}
		p.data = newbuf
	}

	for i := 0; i < len(sendData); i++ {
		p.data[p.datalen+i] = sendData[i]
	}
	p.datalen += len(sendData)
	Info("conn Writeed ", len(sendData), p.datalen, p.bufflen)
	if sendch {

		Debug("write ch ", p)
		p.ch <- 1
	}

	return len(sendData), nil

}

func (p *MockConn) Read(b []byte) (n int, err error) {

	p.lock.Lock()

	if p.datalen == 0 && p.closed != 0 {
		p.lock.Unlock()
		panic("closed ")
		return -1, errors.New("closed")
	}

	if p.datalen == 0 {
		p.lock.Unlock()
		Debug("read ch ", p)
		_ = <-p.ch

		Debug("readed ch ", p)
		p.lock.Lock()
	}
	readlen := 0
	if len(b) < p.datalen {
		readlen = len(b)
	} else {
		readlen = p.datalen
	}

	for i := 0; i < readlen; i++ {
		b[i] = p.data[i]
	}
	for i := 0; i < p.datalen-readlen; i++ {
		p.data[i] = p.data[readlen+i]
	}
	orglen := p.datalen
	p.datalen = p.datalen - readlen
	Info("conn Readed ", readlen, "bytes", "buff ", orglen, p.datalen)

	p.lock.Unlock()
	return readlen, nil
}

func (p *MockConn) Close() error {

	p.lock.Lock()
	defer p.lock.Unlock()

	if p.closed != 0 {
		return errors.New("closed")

	}
	return nil

}
