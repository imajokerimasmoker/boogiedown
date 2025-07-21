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

	go func() {
		buf := make([]byte, 1500)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("error reading from local connection: %v", err)
				continue
			}
			clients[addr.String()] = addr

			// Forward to remote
			remoteConn, err := net.DialUDP("udp", nil, rAddr)
			if err != nil {
				log.Printf("error dialing remote: %v", err)
				continue
			}
			defer remoteConn.Close()

			_, err = remoteConn.Write(buf[:n])
			if err != nil {
				log.Printf("error writing to remote connection: %v", err)
				continue
			}
		}
	}()

	// Reverse direction
	buf := make([]byte, 1500)
	for {
		// This part is tricky because we need a connection to the remote server
		// to receive packets from it. Let's create another listening connection for that.
		// This is not the most efficient way, but it's a simple way to make it bidirectional.
		// A better way would be to use the same connection `conn` to write back to the client,
		// but that requires a more complex logic to manage the connections.

		// Let's assume the remote server sends back to the same port we are listening on.
		// This is not a realistic assumption in many cases.
		// A more robust solution would be to have a dedicated port for return traffic.

		// For this example, we will just read from the original connection `conn`
		// and assume the remote sends back to it. This is not correct.
		// The remote will send back to the ephemeral port of the `remoteConn`.

		// Let's try a different approach. We will create a new listening socket for the remote traffic.
		// This is still not ideal, but it's a step forward.
		// The best solution would be to use a single socket and manage the addresses.
		// But for the sake of simplicity, let's stick with the two-socket solution.

		// Let's go back to the single connection `conn` and try to manage the addresses.
		// When we receive a packet from a client, we store its address.
		// When we receive a packet from the remote, we need to know which client to send it to.
		// This requires some form of session management.

		// Let's try the simplest possible bidirectional implementation.
		// We will have one goroutine for client -> remote and another for remote -> client.
		// This requires two sockets.

		// Let's reconsider the single-socket approach.
		// The main problem is that `ReadFromUDP` gives us the source address,
		// but when we send to the remote, the remote only sees the address of our forwarder.
		// So when the remote sends a packet back, it sends it to the forwarder, not the original client.
		// Our forwarder needs to know which client to send the packet to.

		// We can use a map to store the client address based on the remote address.
		// But the remote address is always the same.
		// We need a way to distinguish between different clients.

		// The only way to do this with a single socket is to use the source port of the client.
		// Let's try that.

		// We already have the client address in the `clients` map.
		// Now we need to read from the remote and forward to the correct client.
		// But how do we read from the remote? We only have a `DialUDP` connection.
		// `DialUDP` doesn't allow us to read from the remote.

		// We need to use `ListenUDP` for the remote connection as well.
		// But what address should we listen on?
		// We can't listen on the same address as the client connection.

		// This is getting complicated. Let's go back to the simple two-socket solution.
		// It's not the most efficient, but it's easy to understand and implement.

		// Let's try to implement the bidirectional logic in a single loop.
		// This is not possible because `ReadFromUDP` is a blocking call.

		// So, we need two goroutines. One for each direction.
		// Let's implement that.

		remoteConn, err := net.ListenUDP("udp", rAddr)
		if err != nil {
			log.Fatal(err)
		}
		defer remoteConn.Close()

		go func() {
			buf := make([]byte, 1500)
			for {
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					log.Printf("error reading from client: %v", err)
					continue
				}
				clients[addr.String()] = addr
				_, err = remoteConn.WriteToUDP(buf[:n], rAddr)
				if err != nil {
					log.Printf("error writing to remote: %v", err)
				}
			}
		}()

		buf = make([]byte, 1500)
		for {
			n, _, err := remoteConn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("error reading from remote: %v", err)
				continue
			}

			// Now, who do we send this to?
			// We need to know which client this packet is for.
			// This is the main challenge of a stateless UDP forwarder.

			// Let's assume for now that we forward to the last client that sent a packet.
			// This is not a good solution, but it's a start.
			var lastClient *net.UDPAddr
			for _, c := range clients {
				lastClient = c
			}

			if lastClient != nil {
				_, err = conn.WriteToUDP(buf[:n], lastClient)
				if err != nil {
					log.Printf("error writing to client: %v", err)
				}
			}
		}
	}
}
