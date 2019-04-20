package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golang.org/x/sys/windows"
)

// LibAssaun file socket: Attempt to read contents of the target file and connect to a TCP port
func dialAssuan(p string, poll bool) (*overlappedFile, error) {
	pipeConn, err := dialPipe(p, poll)
	if err != nil {
		return nil, err
	}

	var port int
	var nonce [16]byte

	reader := bufio.NewReader(pipeConn)

	// Read the target port number from the first line
	tmp, _, err := reader.ReadLine()
	if err == nil {
		port, err = strconv.Atoi(string(tmp))
	}
	if err != nil {
		return nil, err
	}

	// Read the rest of the nonce from the file
	n, err := reader.Read(nonce[:])
	if err != nil {
		return nil, err
	}

	if n != 16 {
		err = fmt.Errorf("Read incorrect number of bytes for nonce. Expected 16, got %d (0x%X)", n, nonce)
		return nil, err
	}

	if *verbose {
		log.Printf("Port: %d, Nonce: %X", port, nonce)
	}

	pipeConn.Close()

	for {
		// Try to connect to the libassaun TCP socket hosted on localhost
		conn, err := dialPort(port, poll)

		if poll && (err == windows.WSAETIMEDOUT || err == windows.WSAECONNREFUSED || err == windows.WSAENETUNREACH || err == windows.ERROR_CONNECTION_REFUSED) {
			time.Sleep(pollTimeout)
			continue
		}

		if err != nil {
			err = os.NewSyscallError("ConnectEx", err)
			return nil, err
		}

		_, err = conn.Write(nonce[:])
		if err != nil {
			return nil, err
		}

		return conn, nil
	}
}
