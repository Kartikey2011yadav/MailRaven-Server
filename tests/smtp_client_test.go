package tests

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

// MockSMTPServer is a simple SMTP server for testing
type MockSMTPServer struct {
	Listener net.Listener
	Port     string
	Messages []string // Store received messages
	mu       sync.Mutex
}

func NewMockSMTPServer() (*MockSMTPServer, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	_, port, _ := net.SplitHostPort(l.Addr().String())

	s := &MockSMTPServer{
		Listener: l,
		Port:     port,
	}

	go s.serve()
	return s, nil
}

func (s *MockSMTPServer) serve() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *MockSMTPServer) handle(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Send greeting
	fmt.Fprintf(writer, "220 localhost ESMTP MockServer\r\n")
	writer.Flush()

	var messageBody strings.Builder
	inData := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimRight(line, "\r\n")

		if inData {
			if line == "." {
				inData = false
				s.addMessage(messageBody.String())
				fmt.Fprintf(writer, "250 OK\r\n")
				writer.Flush()
			} else {
				// Dot un-stuffing
				if strings.HasPrefix(line, ".") {
					line = line[1:]
				}
				messageBody.WriteString(line + "\r\n")
			}
			continue
		}

		cmd := strings.ToUpper(strings.Split(line, " ")[0])
		switch cmd {
		case "EHLO", "HELO":
			// No STARTTLS advertised
			fmt.Fprintf(writer, "250-localhost\r\n250 SIZE 10485760\r\n")
			writer.Flush()
		case "MAIL":
			fmt.Fprintf(writer, "250 OK\r\n")
			writer.Flush()
		case "RCPT":
			fmt.Fprintf(writer, "250 OK\r\n")
			writer.Flush()
		case "DATA":
			inData = true
			fmt.Fprintf(writer, "354 Start mail input; end with <CRLF>.<CRLF>\r\n")
			writer.Flush()
		case "QUIT":
			fmt.Fprintf(writer, "221 Byte\r\n")
			writer.Flush()
			return
		case "STARTTLS":
			fmt.Fprintf(writer, "502 Not implemented\r\n")
			writer.Flush()
		default:
			fmt.Fprintf(writer, "500 Unknown command\r\n")
			writer.Flush()
		}
	}
}

func (s *MockSMTPServer) addMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, msg)
}

func (s *MockSMTPServer) getMessages() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.Messages...)
}

func (s *MockSMTPServer) Close() {
	s.Listener.Close()
}

// TestSMTPClientDelivery tests the SMTP client (T088)
func TestSMTPClientDelivery(t *testing.T) {
	// 1. Start mock server
	mockServer, err := NewMockSMTPServer()
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer mockServer.Close()

	logger := observability.NewLogger("debug", "text")
	client := smtp.NewClient(config.DANEConfig{Mode: "off"}, logger)

	// Configure client to talk to mock server
	client.Port = mockServer.Port

	// Override MX lookup to point to localhost
	client.LookupMX = func(name string) ([]*net.MX, error) {
		return []*net.MX{{Host: "127.0.0.1", Pref: 10}}, nil
	}

	// Test data
	from := "sender@example.com"
	to := "recipient@test.local"
	body := "Subject: Test\r\n\r\nHello World"

	// 2. Perform Send
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Send(ctx, from, to, []byte(body))
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// 3. Verify delivery
	// Wait a bit for async processing if needed (though Send is synchronous up to "250 OK")
	time.Sleep(100 * time.Millisecond)

	// Since MockServer handles connections in goroutines, we need synchronization or sleep.
	// Simplified mock server (above) appends to slice directly - likely racey.
	// But let's see if it works for single test case.
	// Ideally MockSMTPServer should use a mutex.

	if len(mockServer.getMessages()) == 0 {
		// Try waiting a bit more
		time.Sleep(500 * time.Millisecond)
	}

	msgs := mockServer.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message received, got %d", len(msgs))
	}

	received := msgs[0]
	if !strings.Contains(received, "Hello World") {
		t.Errorf("Message body mismatch. Got: %s", received)
	}
}
