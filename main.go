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

	// Create a map to store client addresses
	clients := make(map[string]*net.UDPAddr)
	// Create a channel to signal the main goroutine to exit
	done := make(chan struct{})

	go func() {
		buf := make([]byte, 1500)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("error reading from connection: %v", err)
				continue
			}

			if addr.String() == rAddr.String() {
				// Packet from remote, forward to a client
				// For simplicity, we forward to the first client in the map.
				// A more robust solution would be to track sessions.
				var clientAddr *net.UDPAddr
				for _, c := range clients {
					clientAddr = c
					break
				}
				if clientAddr != nil {
					_, err = conn.WriteToUDP(buf[:n], clientAddr)
					if err != nil {
						log.Printf("error writing to client: %v", err)
					}
				}
			} else {
				// Packet from a client, forward to remote
				clients[addr.String()] = addr
				_, err = conn.WriteToUDP(buf[:n], rAddr)
				if err != nil {
					log.Printf("error writing to remote: %v", err)
				}
			}
		}
	}()

	// Wait for a signal to exit
	<-done
}
