package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

func generateTLSConfig() *tls.Config {
	// Генерируем самоподписанный сертификат для тестирования
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"CloudBridge Test"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatal(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"cloudbridge-p2p"},
	}
}

func main() {
	// Создаем QUIC listener
	addr := "localhost:9091" // Используем другой порт, чтобы не конфликтовать с relay сервером
	
	tlsConfig := generateTLSConfig()
	
	listener, err := quic.ListenAddr(addr, tlsConfig, nil)
	if err != nil {
		log.Fatalf("Failed to create QUIC listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("QUIC test server listening on %s\n", addr)

	for {
		// Принимаем соединение
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		fmt.Printf("New connection from %s\n", conn.RemoteAddr())

		// Обрабатываем соединение в отдельной горутине
		go handleConnection(conn)
	}
}

func handleConnection(conn quic.Connection) {
	defer conn.CloseWithError(0, "server closing")

	// Принимаем потоки
	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Failed to accept stream: %v", err)
			return
		}

		// Обрабатываем поток в отдельной горутине
		go handleStream(stream)
	}
}

func handleStream(stream quic.Stream) {
	defer stream.Close()

	buffer := make([]byte, 4096)
	n, err := stream.Read(buffer)
	if err != nil {
		log.Printf("Failed to read from stream: %v", err)
		return
	}

	message := string(buffer[:n])
	fmt.Printf("Received message: %s\n", message)

	// Отправляем подтверждение
	response := "Message received successfully!"
	_, err = stream.Write([]byte(response))
	if err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}
