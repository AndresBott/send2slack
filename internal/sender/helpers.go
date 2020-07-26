package sender

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

// reads up to max byte size from stdin and returns the content as string
// exits if it did not get any character after some time, this is used to
// determine if it has been invoked as part of a script or manually
func InReader(max int) (string, error) {
	r := bufio.NewReader(os.Stdin)
	s := strings.Builder{}
	initialSignal := false
	ch := make(chan bool, 1)
	chFin := make(chan int, 1)
	byteCount := 0

	go func() {

		b := make([]byte, 8)
		for {
			n, err := r.Read(b)
			byteCount = byteCount + n
			if byteCount >= max {
				chFin <- 1
			}
			if initialSignal == false {
				ch <- true
				initialSignal = true
			}

			s.Write(b[:n])
			if err == io.EOF {
				chFin <- 0
				break
			}
		}

	}()

	select {
	case <-ch:

	case <-time.After(50000000 * time.Nanosecond):
		return "", errors.New("io reader not started")
	}

	fin := <-chFin
	if fin == 1 {
		return "", errors.New("io reader size exceeded")
	}
	return s.String(), nil

}
