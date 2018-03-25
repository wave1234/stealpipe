package conn

import (
	crand "crypto/rand"
	"net"
	"io"
	mrand "math/rand"

	. "github.com/stealpipe/lib/encrypt"
	"testing"
	"time"

	. "github.com/stealpipe/lib/debug"
)

var fakeheaderlen int64
var fakeheaderindex int64

const (
	data_pack_len = 100000
	rand_range    = 100
)

func demo(t *testing.T, socktype int) {

	ln, err := net.Listen("tcp", ":18081")
	if err != nil {
		panic(err)
		t.Fail()
	}
	c := make(chan int)

	s2 := mrand.NewSource(time.Now().Unix())
	r2 := mrand.New(s2)
	fakeheaderlen = int64(r2.Intn(10000)) + 500
	fakeheaderindex = int64(r2.Intn(500))
	datalen := r2.Intn(rand_range) * data_pack_len
	Info("demo datalen:", datalen, socktype)
	data := make([]byte, datalen)
	key := make([]byte, KEYLENGTH)
	if _, err := io.ReadFull(crand.Reader, data); err != nil {
		panic(err)
	}

	data2 := make([]byte, datalen)
	for i := 0; i < datalen; i++ {
		data2[i] = data[i]
	}
	go aes_ctr_c(data2, key, c, socktype)
	Debug("accept")
	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}

	Debug("do server")
	var p ConnInterface

	if socktype == 1 {
		p = &Aes256CTRSocket{}
	} else if socktype == 2 {
		p = &Aes256CBCSocket{}
	} else if socktype == 3 {
		p = &HttpSocket{}
	} else if socktype == 4 {
		p = &TCPSocket{}
	}

	p.SetFakeHeaderLength(fakeheaderlen)

	p.SetFakeHeaderPaddingIndex(fakeheaderindex)

	p.Init(conn, key)
	Debug("server  begin readyRead")
	b := p.ReadyRead()
	if !b {
		t.Fail()
	}
	Debug("server ReadRead Done")
	Debug("Server Begin Read", datalen)
	b, Readbuf := p.Readn(datalen)
	if !b {
		t.Fail()
	}
	Debug("Server Readed", datalen)
	for i := 0; i < datalen; i++ {
		if data[i] != Readbuf[i] {
			t.Fail()
		}
	}

	b = p.ReadyWrite()
	if !b {
		t.Fail()
	}

	Debug("server Write", len(data))
	b = p.Write(data, len(data))
	if !b {
		t.Fail()
	}
	_ = <-c
	conn.Close()
	ln.Close()
	Info("demo done datalen:", datalen, socktype)

}

func Test_aes_ctr(t *testing.T) {

	for i := 0; i < 1; i++ {

		demopipe(t, 1)

		demopipe(t, 2)

		HackGetRandomMode = chunk_Mode
		demopipe(t, 3)

		HackGetRandomMode = content_Mode

		demopipe(t, 3)
		demopipe(t, 4)

		demo(t, 1)
		demo(t, 2)
		HackGetRandomMode = chunk_Mode
		demo(t, 3)
		HackGetRandomMode = content_Mode
		demo(t, 3)
		demo(t, 4)

	}
}

func aes_ctr_c(data []byte, key []byte, c chan int, socktype int) {
	Debug("aec_ctr")
	conn, err := net.Dial("tcp", "127.0.0.1:18081")
	if err != nil {
		panic(err)
	}
	var p ConnInterface

	if socktype == 1 {
		p = &Aes256CTRSocket{}
	} else if socktype == 2 {
		p = &Aes256CBCSocket{}
	} else if socktype == 3 {
		p = &HttpSocket{}
	} else if socktype == 4 {
		p = &TCPSocket{}
	}

	p.SetFakeHeaderLength(fakeheaderlen)
	p.SetFakeHeaderPaddingIndex(fakeheaderindex)

	p.Init(conn, key)
	Debug("do client")
	b := p.ReadyWrite()
	if !b {
		panic(b)
	}

	Debug("client ReadyWrite Done")

	Debug("client write Data", len(data))

	b = p.Write(data, len(data))
	if !b {
		panic(b)
	}

	Debug("client write Data done", len(data))

	Debug("client begin readyRead")
	b = p.ReadyRead()
	if !b {
		panic(b)
	}

	Debug("client readyRead Done")

	Debug("client begin read", len(data))

	b, Readbuf := p.Readn(len(data))
	if !b {
		panic(b)
	}

	Debug("client read Done", len(data))
	for i := 0; i < len(data); i++ {

		if data[i] != Readbuf[i] {
			panic(b)
		}
	}
	p.Close()
	c <- 1
}

func demopipe(t *testing.T, socktype int) {
	Init()

	mkp := MakePipe{}
	mkp.Init()
	Debug(&mkp)
	sConn := mkp.GetServer()
	cConn := mkp.GetClient()
	c := make(chan int)
	s2 := mrand.NewSource(time.Now().Unix())
	r2 := mrand.New(s2)
	fakeheaderlen = int64(10000)
	fakeheaderindex = int64(4000)

	datalen := r2.Intn(rand_range) * data_pack_len
	datalen = 30 * data_pack_len
	Info("demopipe datalen:", datalen, socktype)
	data := make([]byte, datalen)
	key := make([]byte, KEYLENGTH)

	if _, err := io.ReadFull(crand.Reader, data); err != nil {
		panic(err)
	}

	data2 := make([]byte, datalen)
	for i := 0; i < datalen; i++ {
		data2[i] = data[i]
	}
	go aes_ctr_c_pipe(cConn, data2, key, c, socktype)
	Debug("accept")

	Debug("do server")
	var p ConnInterface

	if socktype == 1 {
		p = &Aes256CTRSocket{}
	} else if socktype == 2 {
		p = &Aes256CBCSocket{}
	} else if socktype == 3 {
		p = &HttpSocket{}
	} else if socktype == 4 {
		p = &TCPSocket{}
	}

	p.SetFakeHeaderLength(fakeheaderlen)

	p.SetFakeHeaderPaddingIndex(fakeheaderindex)

	p.Init(sConn, key)
	Debug("server  begin readyRead")
	b := p.ReadyRead()
	if !b {
		t.Fail()
	}
	Debug("server ReadRead Done")
	Debug("Server Begin Read", datalen)
	b, Readbuf := p.Readn(datalen)
	if !b {
		t.Fail()
	}
	Debug("Server Readed", datalen, " Readbuf len", len(Readbuf))
	for i := 0; i < datalen; i++ {
		if data[i] != Readbuf[i] {
			t.Fail()
		}
	}

	b = p.ReadyWrite()
	if !b {
		t.Fail()
	}

	Debug("server Write", len(data))
	b = p.Write(data, len(data))
	if !b {
		t.Fail()
	}
	_ = <-c
	sConn.Close()
	Info("demopipe done datalen:", datalen, socktype)

}

func aes_ctr_c_pipe(conn Pipe, data []byte, key []byte, c chan int, socktype int) {
	Debug("aec_ctr")
	var p ConnInterface

	if socktype == 1 {
		p = &Aes256CTRSocket{}
	} else if socktype == 2 {
		p = &Aes256CBCSocket{}
	} else if socktype == 3 {
		p = &HttpSocket{}
	} else if socktype == 4 {
		p = &TCPSocket{}
	}
	p.SetFakeHeaderLength(fakeheaderlen)

	p.SetFakeHeaderPaddingIndex(fakeheaderindex)

	p.Init(conn, key)
	Debug("do client")
	b := p.ReadyWrite()
	if !b {
		panic(b)
	}

	Debug("client ReadyWrite Done")

	Debug("client write Data", len(data))

	b = p.Write(data, len(data))
	if !b {
		panic(b)
	}

	Debug("client write Data done", len(data))

	Debug("client begin readyRead")
	b = p.ReadyRead()
	if !b {
		panic(b)
	}

	Debug("client readyRead Done")

	Debug("client begin read", len(data))

	b, Readbuf := p.Readn(len(data))
	if !b {
		panic(b)
	}

	Debug("client read Done", len(data))
	for i := 0; i < len(data); i++ {

		if data[i] != Readbuf[i] {
			Info("index ==", i)
			panic(b)
		}
	}
	p.Close()
	c <- 1
}
