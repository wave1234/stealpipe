package quic

import (
	"net"
	"sync"

    )

type connection interface {
	Write([]byte) error
	Read([]byte) (int, net.Addr, error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetCurrentRemoteAddr(net.Addr)
}

type conn struct {
	mutex sync.RWMutex

	 conv *Convertor


	 pconn       net.PacketConn
	currentAddr net.Addr
}

var _ connection = &conn{}

func (c *conn) Write(p []byte) error {

	p2 := c.conv.Encoder(p)

	_, err := c.pconn.WriteTo(p2, c.currentAddr)
	return err
    }

func (c *conn) Read(p []byte) (int, net.Addr, error) {

    l, addr, err := c.pconn.ReadFrom(p)
    l = c.conv.Decoder(p, l)

	  //fmt.Println("read l", l)
    return l,addr, err



    /*
    if l == 0 || (l % aes.BlockSize > 0) {
    fmt.Println("find read l", l)
        return l, addr, err
    }
    b1 := GetDecryptBlockMode(c.AESIv[:], c.AESKey[:])
	b2, p2, _ := Aes256CBCDecrypt(p, l,b1)
	if !b2 {
        panic(b2)
    }
    for i := 0; i < len(p2); i++ {
        p[i]= p2[i]
    }
    for i := len(p2); i< len(p); i++ {
        p[i] = 0
    }
    
    fmt.Println("find read l", len(p2))
    return len(p2), addr, err

	//return c.pconn.ReadFrom(p2)
*/
}

func (c *conn) SetCurrentRemoteAddr(addr net.Addr) {
	c.mutex.Lock()
	c.currentAddr = addr
	c.mutex.Unlock()
}

func (c *conn) LocalAddr() net.Addr {
	return c.pconn.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	c.mutex.RLock()
	addr := c.currentAddr
	c.mutex.RUnlock()
	return addr
}

func (c *conn) Close() error {
	return c.pconn.Close()
}


