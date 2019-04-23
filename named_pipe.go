package main

import (
	"net"
	"os"
	"time"

	"github.com/Microsoft/go-winio"
)

func dialPipe(p string, poll bool) (net.Conn, error) {
	for {
		conn, err := winio.DialPipe(p, nil)
		if err == nil {
			return conn, nil
		} else if poll && os.IsNotExist(err) {
			time.Sleep(pollTimeout)
			continue
		}

		return nil, err
	}
}
