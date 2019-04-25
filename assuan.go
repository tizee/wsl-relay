package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func openFile(fn string, poll bool) (io.Reader, error) {
	for {
		f, err := os.Open(fn)
		if err == nil {
			return f, nil
		} else if poll && (os.IsNotExist(err)) {
			time.Sleep(pollTimeout)
			continue
		}

		return nil, err
	}
}

// LibAssaun file socket: Attempt to read contents of the target file and connect to a TCP port
func dialAssuan(fn string, poll bool) (net.Conn, error) {
	f, err := openFile(fn, poll)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var port int
	var nonce [16]byte

	reader := bytes.NewBuffer(data)

	// Read the target port number from the first line
	tmp, err := reader.ReadString('\n')
	if err == nil {
		// Sanity check, make sure this is actually an int
		port, err = strconv.Atoi(strings.TrimSpace(tmp))
	}
	if err != nil {
		return nil, err
	}

	// Read the rest of the nonce from the file
	n, err := reader.Read(nonce[:])
	if err != nil {
		return nil, err
	} else if n != 16 {
		err = fmt.Errorf("Read incorrect number of bytes for nonce. Expected 16, got %d (0x%X)", n, nonce)
		return nil, err
	}

	if *verbose {
		log.Printf("Port: %d, Nonce: %X", port, nonce)
	}

	for {
		// Try to connect to the libassaun TCP socket hosted on localhost
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)))

		if poll && (err == windows.WSAETIMEDOUT || err == windows.WSAECONNREFUSED || err == windows.WSAENETUNREACH || err == windows.ERROR_CONNECTION_REFUSED) {
			time.Sleep(pollTimeout)
			continue
		}

		if err != nil {
			return nil, err
		}

		_, err = conn.Write(nonce[:])
		if err != nil {
			return nil, err
		}

		return conn, nil
	}
}
