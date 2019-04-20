package main

import (
	"errors"

	"golang.org/x/sys/windows"
)

func dialPort(p int, poll bool) (*overlappedFile, error) {
	if p < 0 || p > 65535 {
		return nil, errors.New("Invalid port value")
	}

	h, err := windows.Socket(windows.AF_INET, windows.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	// Connect to 127.0.0.1
	sa := &windows.SockaddrInet4{Addr: [4]byte{0x7F, 0x00, 0x00, 0x01}, Port: p}

	// Bind to a randomly assigned local port
	err = windows.Bind(h, &windows.SockaddrInet4{})
	if err != nil {
		return nil, err
	}

	// Wrap our socket up to be properly handled
	conn := newOverlappedFile(h)

	// Connect to the socket using overlapped ConnectEx operation
	_, err = conn.asyncIo(func(h windows.Handle, n *uint32, o *windows.Overlapped) error {
		return windows.ConnectEx(h, sa, nil, 0, nil, o)
	})
	if err != nil {
		return nil, err
	}

	return conn, nil
}
