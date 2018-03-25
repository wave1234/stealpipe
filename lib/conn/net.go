package conn

import (
	. "github.com/stealpipe/lib/debug"
	"net"
	"os"
)

func CheckErr(err error) {
	if err != nil {
		Debug(os.Stderr, "Fatal error: ", err.Error())
		os.Exit(-1)
	}
}

func PrintArray(t []byte, x int) {
	for i := 0; i < x; i++ {
	;
	}

}

func SendData(conn net.Conn, sendData []byte) (int, error) {

	n, err := conn.Write(sendData)
	if err != nil {
		return -1, err
	}
	Debug("sended Data ", n)
	return len(sendData), nil
}

func ReadByte(ClientConn net.Conn, needread int, data []byte) bool {

	if needread == 0 {
		return true
	}

	read := 0

	if data == nil {
		Debug("read data ,data == nil", needread)
		data := make([]byte, 1024)
		for {
			x := needread - read
			if x > 1024 {
				x = 1024
			}
			i, err := ClientConn.Read(data[0:x])
			if err != nil {
				return false
			}
			read += i
			if read == needread {
				return true
			}
		}
	}

	for {
		i, err := ClientConn.Read(data[read:needread])
		if err != nil {

			return false
		}
		read += i
		if read == needread {
			return true
		}
	}
}

func Connect(host string, port string) (bool, net.Conn) {
	address, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		Debug("Resolution error", err.Error())
		return false, nil
	}
	client := address.String() + ":" + port
	tcpAddress, err := net.ResolveTCPAddr("tcp4", client)
	if err != nil {
		return false, nil
	}
	Server, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		return false, nil
	}
	return true, Server
}
