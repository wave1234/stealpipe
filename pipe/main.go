package main

/*
   Steal Pipe  is a means of securely transferring computer data between a local host and a remote host or between two remote hosts.
   author: sha512 cbb02d8fa07c171c0fab947010005588f69fa7fede8c4cb4289a8ffc4e244ae72e082e8d68963b3fe8d6dd08e87b75df22c3bbd924873d79ea77dc2c8d29e2ab 
*/
import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	. "github.com/stealpipe/lib/conn"
	. "github.com/stealpipe/lib/encrypt"
	. "github.com/stealpipe/lib/version"

	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"

	quic "github.com/lucas-clemente/quic-go"
)

const (
	PIPECLIENT = "client"
	PIPESERVER = "server"

	CBC  = "CBC"
	CTR  = "CTR"
	HTTP = "HTTP"
	QUIC = "QUIC"
)

var host = flag.String("host", "0.0.0.0", "host 本地IP地址")
var port = flag.String("port", "8721", "port 本地IP端口, 也可以设置为 1-200，这样直接监听多个端口")
var connectionTimeOut = flag.Int("timeout", 500, "connection timeout 连接超时")
var remotehost = flag.String("remotehost", "127.0.0.1:5555", "remotehost:port  远端服务器IP:端口 或者用 127.0.0.1:5555-6666 这样随机使用远端多个服务器端口")

var piptype = flag.String("pipetype", PIPESERVER, "piptype   "+PIPECLIENT+"  or "+PIPESERVER+" 作为客户端 client 还是服务器端 server运行")
var key = flag.String("key", "abc123", "encrypt key  加密密钥")
var encrypttype = flag.String("encrypttype", "CBC", "encrypt type HTTP or CBC QUIC，可以使用的加密办法 HTTP 或 CBC 或 QUIC")
var randomID = flag.Int("r", 1531, "randomID   加密随机数")
var randomSeed = flag.Int("s", 1210, "randomSeed 加密随机因子,必须比加密随机数小")

var LocalHostList []string
var RemoteHostList []string

func GetRandomRemote() string {
	i := mrand.Intn(len(RemoteHostList))
	return RemoteHostList[i]
}

func CheckPort(port string) (bool, int, int) {

	v := strings.Split(port, "-")
	if len(v) != 2 {
		log.Print("port ", port, " format error")
		return false, 0, 0
	}
	iport1, err := strconv.Atoi(v[0])
	if err != nil {
		log.Print("port1 ", v[0], " format error")
		return false, 0, 0
	}
	iport2, err := strconv.Atoi(v[1])
	if err != nil {
		log.Print("port2 ", v[1], " format error")
		return false, 0, 0
	}
	if iport1 > iport2 || (iport1 <= 0) || (iport2 <= 0) || iport2 > 65535 {
		log.Print("port ", *host, " format error")
	}
	return true, iport1, iport2
}

func CheckArgument() bool {

	LocalHostList = make([]string, 0)
	RemoteHostList = make([]string, 0)

	ip := net.ParseIP(*host)
	if ip == nil {
		log.Print("host ", *host, " is not a IP")
		return false
	}
	_, err := strconv.Atoi(*port)
	log.Print("port ", *port)
	if err == nil {
		LocalHostList = append(LocalHostList, *host+":"+*port)
	} else {
		b, iport1, iport2 := CheckPort(*port)
		if !b {
			log.Print("not a right port")
			return false
		}
		for i := iport1; i <= iport2; i++ {
			log.Print("port ", i)
			LocalHostList = append(LocalHostList, *host+":"+strconv.Itoa(i))
		}
	}

	_, err = net.ResolveTCPAddr("tcp", *remotehost)
	if err == nil {
		RemoteHostList = append(RemoteHostList, *remotehost)
	} else {
		log.Print("remotehost ", *remotehost)
		v := strings.Split(*remotehost, ":")
		if len(v) != 2 {
			log.Print("remotehost ", *remotehost, " format error, it should seems like 127.0.0.1:100-200")
			return false
		}
		ip := net.ParseIP(v[0])
		log.Print("ip ", v[0], v[1])
		if ip == nil {
			log.Print("remotehost ", v[0], " is not a IP")
			return false
		}
		localhost := v[0]

		b, iport1, iport2 := CheckPort(v[1])
		if !b {
			return false
		}
		for i := iport1; i <= iport2; i++ {
			log.Print("remote port ", localhost, i)
			RemoteHostList = append(RemoteHostList, localhost+":"+strconv.Itoa(i))
		}
	}

	return true
}

func main() {

	fmt.Println("Secure Pipe version:", Version)
	if len(os.Args) == 1 {
		useage()
		return
	}

	flag.Parse()
	if !CheckArgument() {
		return
	}

	if *encrypttype == QUIC {
		QuicMode()
	} else {
		for _, i := range LocalHostList {

			go HandleOneLocalPort(i)
		}
	}

	for {
		time.Sleep(1 * time.Second)
	}

}

func HandleOneLocalPort(localhost string) {

	tcpAddr, err1 := net.ResolveTCPAddr("tcp", localhost)
	if err1 != nil {
		return
	}

	listen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Print("Error listening:", err)
		return
	}
	defer listen.Close()
	log.Print("Listening on " + localhost)
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			log.Print("Error accepting: ", err)
			return
		}
		log.Printf("Received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())
		conn.SetNoDelay(true)
		go HandleRequest(conn)
	}

}

func HandleRequest(conn *net.TCPConn) {

	defer conn.Close()
	remotehost := GetRandomRemote()
	log.Print("try connect ", remotehost)
	tcpAddr, err := net.ResolveTCPAddr("tcp", remotehost)
	if err != nil {
		log.Print(tcpAddr)
		return
	}

	timeOut := time.Duration(*connectionTimeOut) * time.Millisecond
	ClientConn, err2 := net.DialTimeout("tcp", remotehost, timeOut)
	if err2 != nil {
		log.Print("Error connecting:", err2)
		return
	}
	tcpconn := ClientConn.(*net.TCPConn)
	tcpconn.SetNoDelay(true)

	defer ClientConn.Close()

	var client ConnInterface
	var server ConnInterface

	if *encrypttype == CTR {
		client = &Aes256CTRSocket{}
	} else if *encrypttype == CBC {
		client = &Aes256CBCSocket{}
	} else if *encrypttype == HTTP {
		client = &HttpSocket{}
	}

	server = &TCPSocket{}

	if *piptype == PIPECLIENT {
		client.Init(ClientConn, GetFullKey(*key))
		server.Init(conn, GetFullKey(*key))
	} else {
		client.Init(conn, GetFullKey(*key))
		server.Init(ClientConn, GetFullKey(*key))
	}

	client.SetFakeHeaderLength(int64(*randomID))
	client.SetFakeHeaderPaddingIndex(int64(*randomSeed))

	b := client.ReadyWrite()
	if !b {
		return
	}
	b = client.ReadyRead()
	if !b {
		return
	}
	go Transfer(server, client)
	Transfer(client, server)

}

func Transfer(client ConnInterface, server ConnInterface) {
	defer client.Close()
	defer server.Close()
	sumRead := 0
	sumWrite := 0
	dataLength := 0
	for {
		err, buffer := client.Read()
		if err != true {
			log.Printf("close socket %s <-> %s  r: %d ", client.RemoteAddr(), client.LocalAddr(), sumRead)
			log.Printf("close socket %s <-> %s  w: %d", server.RemoteAddr(), server.LocalAddr(), sumWrite)
			return
		}
		dataLength = len(buffer)
		sumRead += dataLength
		err2 := server.Write(buffer, dataLength)
		if err2 != true {
			return
		}
		sumWrite += dataLength
	}

}

func useage() {
	fmt.Print("Steal Pipe ！！\n")
	fmt.Print("Steal Pipe 是一款开源的安全软件，它可以保护你的数据传输，让它无法被黑客查看和监控.\n")
	fmt.Print("            同时它可以帮助你把本地端口和远端服务器端口连接在一起.\n")
	fmt.Print("\n")

	fmt.Print("可以实现以下功能\n")
	fmt.Print("1. 2个服务器之间进行数据传输， 可以使用CBC HTTP QUIC 3种方式进行加密，其中CBC和HTTP利用TCP, QUIC 使用UDP,使用QUIC的优点是不会被RST影响\n")
	fmt.Print("2. 在服务器上开启隐秘的端口，比如socks5 http服务。 比如 服务器上开启 http 服务，侦听127.0.0.1:80, 开启socks5 服务，侦听127.0.0.1:1080 端口，同时启动PIPE 服务，侦听 7777号UDP端口。\n客户端可以使用PIPE使用服务器的socks5服务，同时也能访问http服务器。服务器的http服务，如果没有PIPE的密码是无法访问的。同时也不会被外部的扫描工具扫描到，是非常安全的。\n\n\n")
	fmt.Print("简单的用法 如下\n")
	fmt.Print("用法1： 本地和远方服务器进行数据传输\n")
	fmt.Print("    本地ip： 192.168.1.1 本地使用8888端口. 远端服务器ip：10.0.0.1，远端服务器使用800-1000端口接受数据. 远端运行了web服务器 服务端口127.0.0.1:80 \n")
	fmt.Print("    本地启动： pipe --key st1234 -r 1000 -s 300 --remotehost  10.0.0.1:800-1000  -port 8888 --pipetype client --encrypttype QUIC\n")
	fmt.Print("    远端服务器启动： pipe --key st1234 -r 1000 -s 300 --remotehost  -port 800-1000 --remotehost 127.0.0.1:80 --encrypttype QUIC\n")

	fmt.Print("\n")

	fmt.Print("用法2： 本地和远方服务器穿越本地HTTP防火墙进行数据传输\n")
	fmt.Print("    本地ip： 192.168.1.1 本地使用8888， 远端服务器ip：10.0.0.1 远端服务器使用1000端口接受数据. 远端运行了web服务器 服务端口 127.0.0.1:80 \n")
	fmt.Print("    本地启动进程1： pipe --remotehost  127.0.0.1:1234  -port 8888 --pipetype client\n")
	fmt.Print("    本地启动进程2： pipe --remotehost  10.0.0.1:1000 --host 127.0.0.1 -port 1234 --pipetype client --encrypttype HTTP \n")

}

func QuicMode() {
	if *piptype == PIPECLIENT {
		QuicClient()
	}
	QuicServer()
}

func QuicClient() {
	for _, host := range LocalHostList {
		go RunOneQuicClient(host)
	}
}

func RunOneQuicClient(localhost string) {
	tcpAddr, err1 := net.ResolveTCPAddr("tcp", localhost)
	if err1 != nil {
		return
	}

	listen, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Print("Error listening:", err)
		return
	}
	defer listen.Close()
	log.Print("Listening on " + localhost)
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			log.Print("Error accepting: ", err)
			return
		}
		log.Printf("Received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())
		conn.SetNoDelay(true)
		go HandleQuicc(conn)
	}
}

func HandleQuicc(conn *net.TCPConn) error {
	defer conn.Close()
	randremotehost := GetRandomRemote()
	config := &quic.Config{}

	s := fmt.Sprintf("%d_%d", *randomID, *randomSeed)
	(*config).AESIv = getKey(s)

	(*config).AESKey = getKey(*key)

	session, err := quic.DialAddr(randremotehost, &tls.Config{InsecureSkipVerify: true}, config)
	if err != nil {
		fmt.Println(err)
		return err
	}

	stream, err := session.OpenStreamSync()
	if err != nil {
		fmt.Println(err)
		return err
	}

	go Transfer2(conn, &stream)
	Transfer3(conn, &stream)

	return nil

}

func Transfer2(client *net.TCPConn, stream *quic.Stream) {
	defer (*stream).Close()
	defer (*client).Close()
	sumRead := 0
	sumWrite := 0
	dataLength := 0
	buff := make([]byte, 1024)
	for {
		l, err := (*stream).Read(buff)
		if err != nil {
			return
		}
		dataLength = l
		sumRead += dataLength
		_, err2 := client.Write(buff[0:dataLength])
		if err2 != nil {
			return
		}
		sumWrite += dataLength
	}
}

func Transfer3(client *net.TCPConn, stream *quic.Stream) {
	defer (*stream).Close()
	defer (*client).Close()
	sumRead := 0
	sumWrite := 0
	dataLength := 0
	buff := make([]byte, 1024)
	for {
		l, err := client.Read(buff)
		if err != nil {
			return
		}
		dataLength = l
		sumRead += dataLength
		_, err2 := (*stream).Write(buff[0:dataLength])
		if err2 != nil {
			return
		}
		sumWrite += dataLength
	}
}

func getKey(t string) [aes.BlockSize]byte {
	h := sha1.New()
	io.WriteString(h, t)
	var r [aes.BlockSize]byte

	copy(h.Sum(nil)[0:aes.BlockSize], r[0:aes.BlockSize])
	return r
}

func QuicServer() {
	fmt.Println("QuicServer()")
	config := &quic.Config{}
	s := fmt.Sprintf("%d_%d", *randomID, *randomSeed)
	(*config).AESIv = getKey(s)

	(*config).AESKey = getKey(*key)

	for _, host := range LocalHostList {
		go RunOneQuicServer(host, config)
	}

}

func RunOneQuicServer(host string, config *quic.Config) {
	log.Printf("RunOneQuicServer ", host)
	listener, err := quic.ListenAddr(host, generateTLSConfig(), config)
	if err != nil {
		panic(err)
		return
	}

	for {
		sess, err := listener.Accept()
		if err != nil {
			continue
		}

		stream, err := sess.AcceptStream()
		if err != nil {
			continue
		}
		go QuicServerOneStream(&stream)
	}
}

func QuicServerOneStream(stream *quic.Stream) {

	randremotehost := GetRandomRemote()

	tcpAddr, err := net.ResolveTCPAddr("tcp", randremotehost)
	if err != nil {
		log.Print(tcpAddr)
		return
	}

	timeOut := time.Duration(*connectionTimeOut) * time.Millisecond
	NetConn, err2 := net.DialTimeout("tcp", randremotehost, timeOut)
	if err2 != nil {
		log.Print("Error connecting:", err2)
		return
	}
	fmt.Println("connect to ", randremotehost)
	tcpConn, ok := NetConn.(*net.TCPConn)
	if !ok {
		return
	}
	go Transfer2(tcpConn, stream)
	Transfer3(tcpConn, stream)

}

// Setup a bare-bones TLS config for the server
// copy from quic-go example

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
