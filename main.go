package main

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	flag "github.com/spf13/pflag"
	"golang.org/x/sys/windows"
)

// How long to sleep between failures while polling
const pollTimeout = 200 * time.Millisecond

var (
	poll            = flag.BoolP("poll", "p", false, "poll until the the specified thing exists")
	closeWrite      = flag.BoolP("close-pipe", "s", false, "close the write channel after stdin closes")
	closeOnEOF      = flag.Bool("pipe-closes", false, "terminate when pipe closes, regardless of stdin state")
	closeOnStdinEOF = flag.Bool("input-closes", false, "terminate on stdin closes, regardless of pipe state")
	verbose         = flag.BoolP("verbose", "v", false, "verbose output on stderr")
	assuan          = flag.Bool("gpg", false, "treat the target as a libassuan file socket (Used by GnuPG)")
)

func underlyingError(err error) error {
	if serr, ok := err.(*os.SyscallError); ok {
		return serr.Err
	}
	return err
}

type closeWriter interface {
	CloseWrite() error
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	var conn net.Conn
	var err error

	if !*assuan {
		if *verbose {
			log.Println("Creating a pipe to", args[0])
		}
		conn, err = dialPipe(args[0], *poll)
	} else {
		if *verbose {
			log.Println("Opening an Assuan connection via", args[0])
		}

		conn, err = dialAssuan(args[0], *poll)
	}

	if err != nil {
		log.Fatalln(err)
	}

	if *verbose {
		log.Println("connected")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			log.Fatalln("copy from stdin to pipe failed:", err)
		}

		if *verbose {
			log.Println("copy from stdin to pipe finished")
		}

		if *closeOnStdinEOF {
			os.Exit(0)
		}

		if *closeWrite {
			cw, ok := conn.(closeWriter)
			if ok {
				err = cw.CloseWrite()
				if err != nil {
					if *verbose {
						log.Println("Failed to close pipe:", err)
					}
				}
			}
		}
		os.Stdin.Close()
		wg.Done()
	}()

	_, err = io.Copy(os.Stdout, conn)
	if underlyingError(err) == windows.ERROR_BROKEN_PIPE || underlyingError(err) == windows.ERROR_PIPE_NOT_CONNECTED {
		// The named pipe is closed and there is no more data to read. Since
		// named pipes are not bidirectional, there is no way for the other side
		// of the pipe to get more data, so do not wait for the stdin copy to
		// finish.
		if *verbose {
			log.Println("copy from pipe to stdout finished: pipe closed")
		}
		os.Exit(0)
	}

	if err != nil {
		log.Fatalln("copy from pipe to stdout failed:", err)
	}

	if *verbose {
		log.Println("copy from pipe to stdout finished")
	}

	if !*closeOnEOF {
		os.Stdout.Close()

		// Keep reading until we get ERROR_BROKEN_PIPE or the copy from stdin
		// finishes.
		go func() {
			for {
				_, err := conn.Read(nil)
				if underlyingError(err) == windows.ERROR_BROKEN_PIPE {
					if *verbose {
						log.Println("pipe closed")
					}
					os.Exit(0)
				} else if err != nil {
					log.Fatalln("pipe error:", err)
				}
			}
		}()

		wg.Wait()
	}
}
