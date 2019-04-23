package grpcgcp

import (
	"context"
	"net"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	defaultTCPUserTimeout = 20000
)

func dialTCPUserTimeout(ctx context.Context, addr string, tcpUserTimeout int) (net.Conn, error) {
	control := func(network, address string, c syscall.RawConn) error {
		var controlErr error
		controlErr = c.Control(func(fd uintptr) {
			controlErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, unix.TCP_USER_TIMEOUT, tcpUserTimeout)
		})
		return controlErr
	}
	d := &net.Dialer{
		Control: control,
	}
	if ctxDeadline, ok := ctx.Deadline(); ok {
		d.Timeout = time.Until(ctxDeadline)
	}
	return d.DialContext(ctx, "tcp", addr)
}

// ContextDialerConfig contains config values for customized context dialer.
type ContextDialerConfig struct {
	// TCPUserTimeout sets the TCP_USER_TIMEOUT socket option in milliseconds.
	// Default is 20000
	TCPUserTimeout uint
}

// NewContextDialer returns a customized context dialer that has additional setups before actual dial.
// It currently supports TCP_USER_TIMEOUT socket option.
func NewContextDialer(config ContextDialerConfig) func(context.Context, string) (net.Conn, error) {
	tcpUserTimeout := config.TCPUserTimeout
	if tcpUserTimeout == 0 {
		tcpUserTimeout = defaultTCPUserTimeout
	}
	return func(ctx context.Context, addr string) (net.Conn, error) {
		return dialTCPUserTimeout(ctx, addr, int(tcpUserTimeout))
	}
}
