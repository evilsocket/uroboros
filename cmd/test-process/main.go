package main

import (
	"bufio"
	"fmt"
	"github.com/evilsocket/islazy/str"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

var period = 3 * time.Second
var stop = false
var randDelay = false
var done = sync.WaitGroup{}
var router = chi.NewRouter()

func fibonacci(n int) int {
	if n <= 1 || stop {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func delay() {
	if randDelay {
		time.Sleep(time.Duration(3000*rand.Float32()) * time.Millisecond)
	}
}

func stressCPU() {
	defer done.Done()
	fmt.Printf("CPU stress running (rand delay %v)...\n", randDelay)
	for !stop {
		fibonacci(45)
		delay()
	}
	fmt.Println("CPU stress finished")
}

func stressMEM() {
	defer done.Done()
	fmt.Printf("MEM stress running (rand delay %v)...\n", randDelay)
	all := [][]int{}

	for !stop {
		for i := 0; i < 300000 && !stop; i++ {
			delay()
			arr := make([]int, 10)
			for j := range arr {
				arr[j] = 0xff
			}
			all = append(all, arr)
		}
	}
	all = nil
	fmt.Println("MEM stress finished")
}

func main() {
	router.Mount("/debug", middleware.Profiler())

	go func() {
		fmt.Println("/debug api on :8080\n")
		panic(http.ListenAndServe("0.0.0.0:8080", router))
	}()

	fmt.Printf("pid: %d\n", os.Getpid())

	reader := bufio.NewReader(os.Stdin)

	for {
		wait := true
		stop = false
		randDelay = false

		fmt.Printf("cpu(r), mem(r), gc: ")
		text, _ := reader.ReadString('\n')
		text = str.Trim(text)

		if text == "cpu" || text == "cpur" {
			randDelay = text == "cpur"
			done.Add(1)
			go stressCPU()
		} else if text == "mem" || text == "memr" {
			randDelay = text == "memr"
			done.Add(1)
			go stressMEM()
		} else if text == "gc" {
			fmt.Println("calling gc ...")
			runtime.GC()
			wait = false
		} else {
			fmt.Printf("unknown command '%s'\n", text)
			continue
		}

		if wait {
			fmt.Println("press any key and enter to stop ...")
			_, _ = reader.ReadString('\n')
			stop = true
			done.Wait()
		}
	}
}
