package grpcgcp

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

const (
	network = "tcp"
	address = ":3000"
)

func startServer(t *testing.T) {
	l, err := net.Listen(network, address)
	if err != nil {
		t.Fatalf("error listening: %v", err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		t.Fatalf("error accepting: %v", err)
	}
	defer conn.Close()
}

func getTCPUserTimeout(conn net.Conn) (int, error) {
	tcpconn, ok := conn.(*net.TCPConn)
	if !ok {
		return 0, fmt.Errorf("conn is not *net.TCPConn. got %T", conn)
	}
	rawConn, err := tcpconn.SyscallConn()
	if err != nil {
		return 0, fmt.Errorf("error getting raw connection: %v", err)
	}
	var timeout int
	err = rawConn.Control(func(fd uintptr) {
		timeout, err = syscall.GetsockoptInt(int(fd), syscall.IPPROTO_TCP, unix.TCP_USER_TIMEOUT)
	})
	if err != nil {
		return 0, fmt.Errorf("error getting option on socket: %v", err)
	}
	return timeout, nil
}

func TestContextDialerDefault(t *testing.T) {
	go startServer(t)

	// Wait for server to start
	time.Sleep(2 * time.Second)

	dialer := NewContextDialer(ContextDialerConfig{})
	conn, err := dialer(context.Background(), address)
	if err != nil {
		t.Fatalf("dial failure: %v", err)
	}
	defer conn.Close()

	timeout, err := getTCPUserTimeout(conn)
	if err != nil {
		t.Fatalf("error getting tcp user timeout: %v", err)
	}
	if timeout != defaultTCPUserTimeout {
		t.Fatalf("tcp user timeout should be %v, got %v", defaultTCPUserTimeout, timeout)
	}
}

func TestContextDialerWithConfig(t *testing.T) {
	go startServer(t)

	// Wait for server to start
	time.Sleep(2 * time.Second)

	testTimeout := 30000
	dialer := NewContextDialer(ContextDialerConfig{TCPUserTimeout: uint(testTimeout)})
	conn, err := dialer(context.Background(), address)
	if err != nil {
		t.Fatalf("dial failure: %v", err)
	}
	defer conn.Close()

	timeout, err := getTCPUserTimeout(conn)
	if err != nil {
		t.Fatalf("error getting tcp user timeout: %v", err)
	}
	if timeout != testTimeout {
		t.Fatalf("tcp user timeout should be %v, got %v", testTimeout, timeout)
	}
}
