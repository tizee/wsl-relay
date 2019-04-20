package main

import (
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func dialPipe(p string, poll bool) (*overlappedFile, error) {
	p16, err := windows.UTF16FromString(p)
	if err != nil {
		return nil, err
	}
	for {
		h, err := windows.CreateFile(&p16[0], windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
		if err == nil {
			return newOverlappedFile(h), nil
		}
		if poll && os.IsNotExist(err) {
			time.Sleep(pollTimeout)
			continue
		}
		return nil, &os.PathError{Path: p, Op: "open", Err: err}
	}
}
