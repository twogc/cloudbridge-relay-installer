package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Listen on external port 8082
	listener, err := net.Listen("tcp", "0.0.0.0:8082")
	if err != nil {
		log.Fatalf("Failed to listen on port 8082: %v", err)
	}
	defer listener.Close()

	log.Println("P2P API Proxy listening on :8082 -> 192.168.49.2:30082")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down proxy...")
		listener.Close()
		os.Exit(0)
	}()

	for {
		// Accept incoming connections
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle each connection in a goroutine
		go handleConnection(clientConn)
	}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Connect to Minikube NodePort
	serverConn, err := net.Dial("tcp", "192.168.49.2:30082")
	if err != nil {
		log.Printf("Failed to connect to Minikube: %v", err)
		return
	}
	defer serverConn.Close()

	log.Printf("Proxying connection from %s to Minikube", clientConn.RemoteAddr())

	// Copy data bidirectionally
	go func() {
		io.Copy(serverConn, clientConn)
		serverConn.Close()
	}()

	io.Copy(clientConn, serverConn)
}
