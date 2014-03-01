package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

var opts struct {
	children    int
	bufferLines int
	quiet       bool
}

func main() {
	flag.BoolVar(&opts.quiet, "quiet", false, "Do not warn about nonzero child exit")
	flag.IntVar(&opts.children, "children", 8, "Number of child processes to balance data to")
	flag.IntVar(&opts.bufferLines, "buffer", 10000, "Number of input lines to buffer")
	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	runtime.GOMAXPROCS(runtime.NumCPU())

	if flag.NArg() != 1 {
		log.Println("Usage:\n  para [options] <command>")
		return
	}

	var inbox = make(chan []byte, opts.bufferLines)
	var buffers = make(chan []byte, opts.bufferLines)

	var wg sync.WaitGroup
	wg.Add(opts.children)
	for i := 0; i < opts.children; i++ {
		i := i
		go func() {
			writeToChild(i, flag.Arg(0), inbox, buffers)
			wg.Done()
		}()
	}

	readInput(inbox, buffers)

	wg.Wait()
}

func readInput(inbox, buffers chan []byte) {
	// When we're all done, make sure we close the inbox to signal end of data.
	defer func() {
		close(inbox)
	}()

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {

		// Read one line from stdin, get a slice to it. The underlying data
		// will be overwritten at next Scan() so we need to buffer it.

		tb := s.Bytes()
		tbl := len(tb)

		// Try to get a reused buffer. If we didn't get any or it was to short,
		// allocate a new buffer instead.

		var bb []byte
		select {
		case bb = <-buffers:
		default:
		}
		if len(bb) < tbl+1 {
			bb = make([]byte, tbl+1)
		}

		// Copy the line to the buffer and post it in the inbox. Add the
		// newline that was stripped by Scan().
		copy(bb, tb)
		bb[tbl] = '\n'
		inbox <- bb
	}
}

func writeToChild(index int, shellcmd string, inbox, buffers chan []byte) {
	env := os.Environ()
	env = append(env, fmt.Sprintf("PARAIDX=%d", index))

	cmd := exec.Command("/bin/sh", "-c", shellcmd)
	cmd.Env = env
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	w, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Fatal: %v (cannot create pipe to child)", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatalf("Fatal: %v (cannot start child)", err)
	}

	writeOutput(w, inbox, buffers)

	err = cmd.Wait()
	if !opts.quiet && err != nil {
		log.Printf("Warning: %v (child exited with error)", err)
	}
}

func writeOutput(w io.WriteCloser, inbox, buffers chan []byte) {
	defer w.Close()

	for b := range inbox {
		// Write out each line.
		_, err := w.Write(b)
		if err != nil {
			log.Printf("Warning: %v (stopping output to child)", err)
			return
		}

		// Try to put the buffer back on the queue of buffers to reuse, or just
		// drop it if the queue is full.
		select {
		case buffers <- b:
		default:
		}
	}
}
