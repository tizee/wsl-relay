package main

import (
	"bufio"
	"log"

	"github.com/Microsoft/go-winio"
)

type closeWriter interface {
	CloseWrite() error
}

func main() {
	pc := &winio.PipeConfig{
		MessageMode:      true,
		InputBufferSize:  1024,
		OutputBufferSize: 1024,
	}
	ln, err := winio.ListenPipe("//./pipe/drtest", pc)
	if err != nil {
		log.Fatalln("Failed to listen to ye pipe", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln("Failed to accept ye pipe", err)
		}

		log.Printf("Hello world")

		reader := bufio.NewReader(conn)
		input := make([]byte, 16)
		_, err = reader.Read(input)
		if err != nil {
			log.Fatalln("Failed to read from ye pipe", err)
		}

		log.Printf("Got me a %s!", string(input))

		output := make([]byte, 0, 20)
		output = append(output, []byte("heck")...)
		output = append(output, input...)

		conn.Write(output)
		cw, ok := conn.(closeWriter)
		if ok {
			err = cw.CloseWrite()
			if err != nil {
				log.Println("Failed to close ye pipe:", err)
			}
		} else {
			conn.Close()
		}
	}

}
