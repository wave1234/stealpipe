package socks5

import (
	"fmt"
	"log"
	"net"
)

func IsSocks(buffer []byte) bool {
	if len(buffer) < 3 {
		log.Print("not socks, len < 3", len(buffer))
		return false
	}
	if (buffer[0] == 5) && (buffer[2] == 0) {
		return true
	}
	if (buffer[0] == 4) && (buffer[1] == 1) {
		return true
	}
	log.Print("wrong request", buffer)
	return false
}

func IsSocks4(buffer []byte) bool {

	if len(buffer) >= 9 && (buffer[0] == 4) && (buffer[1] == 1) && buffer[4] != 0 && buffer[5] != 0 && buffer[6] != 0 && buffer[7] != 0 {
		return true
	} else if len(buffer) > 11 && (buffer[0] == 4) && (buffer[1] == 1) && buffer[len(buffer)-1] == 0 && buffer[4] == 0 && buffer[5] == 0 && buffer[6] == 0 && buffer[7] != 0 { // socks4a
		return true
	}
	return false
}

func GetSocks4Add(buffer []byte) (string, string) {
	log.Printf("GetSocks4Add buffer")
	host := ""
	port := ""
	if len(buffer) > 11 && (buffer[0] == 4) && (buffer[1] == 1) && buffer[len(buffer)-1] == 0 && buffer[4] == 0 && buffer[5] == 0 && buffer[6] == 0 && buffer[7] != 0 {
		host = string(buffer[9 : len(buffer)-1])
		port = fmt.Sprintf("%d", (uint(buffer[2])*256 + uint(buffer[3])))
	} else {
		host = fmt.Sprintf("%d.%d.%d.%d", buffer[4], buffer[5], buffer[6], buffer[7])
		port = fmt.Sprintf("%d", (uint(buffer[2])*256 + uint(buffer[3])))
	}
	return host, port
}

func GetSock4FirstResponse(buffer []byte) []byte {
	response := make([]byte, 8)
	copy(response, buffer[0:8])
	response[0] = 0
	response[1] = 90
	return response
}

func IsSock5ShakeHand(buffer []byte) bool {
	if (buffer[0] == 5) && (buffer[2] == 0) {
		return true
	}
	return false
}

func GetSocks5ShakeHandResponse(buffer []byte) []byte {
	response := make([]byte, 2)
	response[0] = 5
	response[1] = 0
	return response
}

func IsSocks5ConnectRequest(buffer []byte) bool {

	if len(buffer) < 7 {
		log.Print("socks5 too short  ", buffer)
		return false
	}

	if (buffer[0] != 5) && (buffer[1] != 1) && (buffer[2] != 0) {
		log.Print("socks5 wrong header ", buffer)
		return false
	}
	if (buffer[3] != 1) && (buffer[3] != 3) {
		log.Print("socks5 unsuport type, it should be 1 or 3 ", buffer[3])
		return false
	}
	bufLength := 0
	if buffer[3] == 3 {
		bufLength = 5 + int(buffer[4]) + 2

	} else if buffer[3] == 1 {
		bufLength = 4 + 4 + 2
	}
	if bufLength != len(buffer) { // it is a bug, but i will not fix it
		return false
	}
	return true
}

func GetSocks5Add(buffer []byte) (string, string) {
	bufLength := 0
	host := ""
	port := ""
	if buffer[3] == 3 {
		bufLength = 5 + int(buffer[4]) + 2

		host = string(buffer[5 : int(buffer[4])+5])

	} else if buffer[3] == 1 {
		bufLength = 4 + 4 + 2
		host = fmt.Sprintf("%d.%d.%d.%d", buffer[4], buffer[5], buffer[6], buffer[7])
	}
	port = fmt.Sprintf("%d", (uint(buffer[bufLength-2])*256 + uint(buffer[bufLength-1])))
	log.Print("socks5  : ", host)
	return host, port
}

func GetSocks5ConnectResponse(buffer []byte) []byte {
	response := make([]byte, len(buffer))
	copy(response, buffer)

	response[0] = 5
	response[1] = 0
	response[2] = 0
	return response
}

func HandleRequest(conn net.Conn) {
	buf := make([]byte, 100)

	len, err := conn.Read(buf)
	if err != nil {
		log.Print("Error reading:", err.Error())
		conn.Close()
		return
	}

	req := buf[0:len]
	if !IsSocks(req) {
		conn.Close()
		return
	}
	if IsSocks4(req) {
		ProcessSocks4(req, conn)

	} else if IsSock5ShakeHand(req) {
		ProcessSocks5(req, conn)
	} else {
		conn.Close()
		return
	}

}

func ProcessSocks4(buff []byte, conn net.Conn) {
	host, port := GetSocks4Add(buff)
	response := GetSock4FirstResponse(buff)
	defer conn.Close()
	buff = nil
	_, err := conn.Write(response)
	if err != nil {
		return
	}

	ClientConn, err2 := net.Dial("tcp", host+":"+port)
	if err2 != nil {
		log.Print("Error connecting:", err2)
		return
	}

	go Transfer(ClientConn, conn, host)

	Transfer(conn, ClientConn, host)

}

func ProcessSocks5(buff []byte, conn net.Conn) {
	defer conn.Close()
	if !IsSock5ShakeHand(buff) {
		return
	}
	response := GetSocks5ShakeHandResponse(buff)

	defer conn.Close()

	_, err := conn.Write(response)
	if err != nil {
		return
	}
	buff2 := make([]byte, 120)
	l2 := 0
	l2, err = conn.Read(buff2)
	if err != nil {
		log.Print("Error reading:", err.Error())

		return
	}
	req := buff2[0:l2]
	if !IsSocks5ConnectRequest(req) {
		return

	}

	host, port := GetSocks5Add(req)
	response2 := GetSocks5ConnectResponse(req)

	_, err = conn.Write(response2)
	if err != nil {
		return
	}

	ClientConn, err2 := net.Dial("tcp", host+":"+port)
	if err2 != nil {
		log.Print("Error connecting:", err2)
		return
	}
	go Transfer(ClientConn, conn, host)
	Transfer(conn, ClientConn, host)
}

func Transfer(conn1 net.Conn, conn2 net.Conn, host string) {
	defer conn1.Close()
	defer conn2.Close()
	sumread := 0
	sumwrite := 0

	buf := make([]byte, 1024)

	for {

		length, err := conn1.Read(buf)
		if err != nil {
			log.Printf("close socket %s %s <-> %s  r: %d ", host, conn1.RemoteAddr(), conn1.LocalAddr(), sumread)
			log.Printf("close socket %s %s <-> %s  w: %d", host, conn2.RemoteAddr(), conn2.LocalAddr(), sumwrite)

			return
		}
		sumread += length
		_, err2 := conn2.Write(buf[0:length])
		if err2 != nil {
			log.Printf("close socket %s %s <-> %s  r: %d ", host, conn1.RemoteAddr(), conn1.LocalAddr(), sumread)
			log.Printf("close socket %s %s <-> %s  w: %d", host, conn2.RemoteAddr(), conn2.LocalAddr(), sumwrite)

			return
		}
		sumwrite += length

	}

}

