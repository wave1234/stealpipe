package main

import (
	"flag"
	. "github.com/stealpipe/lib/socks5"
	"log"
	"net"
)

var host = flag.String("host", "0.0.0.0", "host")
var port = flag.String("port", "1080", "port")

func main() {

	flag.Parse()
	listen, err := net.Listen("tcp", *host+":"+*port)
	if err != nil {
		log.Print("Error listening:", err)
		return
	}
	defer listen.Close()
	log.Print("Listening on " + *host + ":" + *port)
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Print("Error accepting: ", err)
			return
		}
		//logs an incoming message
		log.Printf("Received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())
		// Handle connections in a new goroutine.
		go HandleRequest(conn)
	}

}
