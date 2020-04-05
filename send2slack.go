package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func main() {

	//c1 := make(chan string, 1)
	//go func() {
	//	time.Sleep(2 * time.Second)
	//	c1 <- "result 1"
	//}()
	//
	//	select {
	//	case res := <-c1:
	//		fmt.Println(res)
	//	case <-time.After(1 * time.Second):
	//		fmt.Println("timeout 1")
	//	}
	//
	//	os.Exit(0)

	//cmd.RootCmd()

	r := bufio.NewReader(os.Stdin)

	s := strings.Builder{}
	initialSignal := false
	ch := make(chan bool, 1)
	chFin := make(chan bool, 1)

	go func() {

		b := make([]byte, 16)
		for {
			n, err := r.Read(b)
			if initialSignal == false {
				ch <- true
				initialSignal = true
			}

			s.Write(b[:n])
			if err == io.EOF {
				chFin <- true
				break
			}
		}

	}()

	fmt.Println("")
	fmt.Println("")
	fmt.Println("=====================================")

	select {
	case <-ch:

	case <-time.After(90000000 * time.Nanosecond):
		fmt.Println("timeout 1")
		os.Exit(0)
	}

	<-chFin
	fmt.Println(s.String())

}
