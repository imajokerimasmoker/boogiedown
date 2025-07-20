package main

import (
	"flag"
	"log"
	"net"
)

func main() {
	var (
		localAddr  = flag.String("l", ":5000", "local address")
		remoteAddr = flag.String("r", "127.0.0.1:6000", "remote address")
	)
	flag.Parse()

	log.Printf("UDP forwarder listening on %s and forwarding to %s", *localAddr, *remoteAddr)

	lAddr, err := net.ResolveUDPAddr("udp", *localAddr)
	if err != nil {
		log.Fatal(err)
	}

	rAddr, err := net.ResolveUDPAddr("udp", *remoteAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", lAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create a single connection to the remote address
	remoteConn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer remoteConn.Close()

	// Create a buffer for incoming packets
	buf := make([]byte, 1500)

	for {
		// Read from the local connection
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("error reading from local connection: %v", err)
			continue
		}

		// Write to the remote connection
		_, err = remoteConn.Write(buf[:n])
		if err != nil {
			log.Printf("error writing to remote connection: %v", err)
			continue
		}
	}
}
